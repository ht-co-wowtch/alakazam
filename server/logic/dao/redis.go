package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	log "github.com/golang/glog"
	"github.com/gomodule/redigo/redis"

	"github.com/zhenjl/cityhash"
)

const (
	// user id的前綴詞，用於存儲在redis當key
	_prefixUidInfo = "uid_%s"

	// server name的前綴詞，用於存儲在redis當key
	_prefixServerOnline = "server_%s"

	// user hash table name key
	HashNameKey = "name"

	// user hash table status key
	hashStatusKey = "status"

	// user hash table server key
	hashServerKey = "server"
)

func keyUidInfo(uid string) string {
	return fmt.Sprintf(_prefixUidInfo, uid)
}

func keyServerOnline(key string) string {
	return fmt.Sprintf(_prefixServerOnline, key)
}

// ping redis是否活著
func (d *Dao) pingRedis(c context.Context) (err error) {
	conn := d.redis.Get()
	_, err = conn.Do("SET", "PING", "PONG")
	conn.Close()
	return
}

// 儲存user資訊
// HSET :
// 主key => uid_{user id}
// user key => user roomId
// name => user name
// status => user status
// server => comet server name
func (d *Dao) AddMapping(c context.Context, uid, key, roomId, name, server string, status int) (err error) {
	conn := d.redis.Get()
	defer conn.Close()
	if err = conn.Send("HSET", keyUidInfo(uid), key, roomId, HashNameKey, name, hashStatusKey, status, hashServerKey, server); err != nil {
		log.Errorf("conn.Send(HSET %s,%s) error(%v)", uid, key, err)
		return
	}
	if err = conn.Send("EXPIRE", keyUidInfo(uid), d.redisExpire); err != nil {
		log.Errorf("conn.Send(EXPIRE %s,%s) error(%v)", uid, key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Errorf("conn.Flush() error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Errorf("conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// 更換房間
func (d *Dao) ChangeRoom(c context.Context, uid, key, roomId string) (err error) {
	conn := d.redis.Get()
	defer conn.Close()
	if err = conn.Send("HSET", keyUidInfo(uid), key, roomId); err != nil {
		log.Errorf("conn.Send(HSET %s,%s) error(%v)", uid, key, err)
		return
	}
	if err = conn.Send("EXPIRE", keyUidInfo(uid), d.redisExpire); err != nil {
		log.Errorf("conn.Send(EXPIRE %s,%s) error(%v)", uid, key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Errorf("conn.Flush() error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Errorf("conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// restart user資料的過期時間
// EXPIRE : uid_{user id}  (HSET)
func (d *Dao) ExpireMapping(c context.Context, uid string) (has bool, err error) {
	conn := d.redis.Get()
	defer conn.Close()
	if err = conn.Send("EXPIRE", keyUidInfo(uid), d.redisExpire); err != nil {
		log.Errorf("conn.Send(EXPIRE %s) error(%v)", uid, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Errorf("conn.Flush() error(%v)", err)
		return
	}
	if has, err = redis.Bool(conn.Receive()); err != nil {
		log.Errorf("conn.Receive() error(%v)", err)
		return
	}
	return
}

// 移除user資訊
// DEL : uid_{user id}
func (d *Dao) DelMapping(c context.Context, uid, key, server string) (has bool, err error) {
	conn := d.redis.Get()
	defer conn.Close()
	if err = conn.Send("HDEL", keyUidInfo(uid), key); err != nil {
		log.Errorf("conn.Send(HDEL %s,%s) error(%v)", uid, server, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Errorf("conn.Flush() error(%v)", err)
		return
	}
	if has, err = redis.Bool(conn.Receive()); err != nil {
		log.Errorf("conn.Receive() error(%v)", err)
		return
	}
	return
}

// 取user資料
func (d *Dao) UserData(uid string, key string) (roomId, name string, status int, err error) {
	conn := d.redis.Get()
	defer conn.Close()
	if err = conn.Send("HGETALL", keyUidInfo(uid)); err != nil {
		log.Errorf("conn.Do(HGETALL %s) error(%v)", uid, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Errorf("conn.Flush() error(%v)", err)
		return
	}
	res, err := redis.StringMap(conn.Receive())
	if err != nil {
		log.Errorf("conn.Receive() error(%v)", err)
		return
	}
	roomId = res[key]
	name = res[HashNameKey]

	// TODO 自行實作redis.StringMap
	if s, err := strconv.Atoi(res[hashStatusKey]); err == nil {
		status = s
	}
	return
}

type Online struct {
	Server    string           `json:"server"`
	RoomCount map[string]int32 `json:"room_count"`
	Updated   int64            `json:"updated"`
}

// 以HSET方式儲存房間人數
// HSET Key hashKey jsonBody
// Key用server name
// hashKey則是將room name以City Hash32做hash後得出一個數字，以這個數字當hashKey
// 至於為什麼hashKey還要用City Hash32做hash就不知道
func (d *Dao) AddServerOnline(c context.Context, server string, online *Online) (err error) {
	roomsMap := map[uint32]map[string]int32{}
	for room, count := range online.RoomCount {
		rMap := roomsMap[cityhash.CityHash32([]byte(room), uint32(len(room)))%64]
		if rMap == nil {
			rMap = make(map[string]int32)
			roomsMap[cityhash.CityHash32([]byte(room), uint32(len(room)))%64] = rMap
		}
		rMap[room] = count
	}
	key := keyServerOnline(server)
	for hashKey, value := range roomsMap {
		err = d.addServerOnline(c, key, strconv.FormatInt(int64(hashKey), 10), &Online{RoomCount: value, Server: online.Server, Updated: online.Updated})
		if err != nil {
			return
		}
	}
	return
}

// 以HSET方式儲存房間人數
// HSET Key hashKey jsonBody
// Key用server name
func (d *Dao) addServerOnline(c context.Context, key string, hashKey string, online *Online) (err error) {
	conn := d.redis.Get()
	defer conn.Close()
	b, _ := json.Marshal(online)
	if err = conn.Send("HSET", key, hashKey, b); err != nil {
		log.Errorf("conn.Send(SET %s,%s) error(%v)", key, hashKey, err)
		return
	}
	if err = conn.Send("EXPIRE", key, d.redisExpire); err != nil {
		log.Errorf("conn.Send(EXPIRE %s) error(%v)", key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Errorf("conn.Flush() error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Errorf("conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// 根據server name取線上各房間總人數
func (d *Dao) ServerOnline(c context.Context, server string) (online *Online, err error) {
	online = &Online{RoomCount: map[string]int32{}}
	// server name
	key := keyServerOnline(server)
	for i := 0; i < 64; i++ {
		ol, err := d.serverOnline(c, key, strconv.FormatInt(int64(i), 10))
		if err == nil && ol != nil {
			online.Server = ol.Server
			if ol.Updated > online.Updated {
				online.Updated = ol.Updated
			}
			for room, count := range ol.RoomCount {
				online.RoomCount[room] = count
			}
		}
	}
	return
}

// 根據server name與hashKey取該server name內線上各房間總人數
func (d *Dao) serverOnline(c context.Context, key string, hashKey string) (online *Online, err error) {
	conn := d.redis.Get()
	defer conn.Close()
	// b是一個json
	// {
	// 		"server":"ne0002de-MacBook-Pro.local",
	// 		"room_count":{
	// 			"chat://1000":1
	// 		 },
	// 		 "updated":1556368160
	// }"
	// chat://1000是房間type + id，1是人數
	// updated是資料更新時間
	b, err := redis.Bytes(conn.Do("HGET", key, hashKey))
	if err != nil {
		if err != redis.ErrNil {
			log.Errorf("conn.Do(HGET %s %s) error(%v)", key, hashKey, err)
		}
		return
	}
	online = new(Online)
	if err = json.Unmarshal(b, online); err != nil {
		log.Errorf("serverOnline json.Unmarshal(%s) error(%v)", b, err)
		return
	}
	return
}

// 根據server name 刪除線上各房間總人數
func (d *Dao) DelServerOnline(c context.Context, server string) (err error) {
	conn := d.redis.Get()
	defer conn.Close()
	key := keyServerOnline(server)
	if _, err = conn.Do("DEL", key); err != nil {
		log.Errorf("conn.Do(DEL %s) error(%v)", key, err)
	}
	return
}
