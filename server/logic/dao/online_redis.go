package dao

import (
	"encoding/json"
	log "github.com/golang/glog"
	"github.com/gomodule/redigo/redis"
	"github.com/zhenjl/cityhash"
	"strconv"
)

type Online struct {
	Server    string           `json:"server"`
	RoomCount map[string]int32 `json:"room_count"`
	Updated   int64            `json:"updated"`
}

// 以HSET方式儲存房間人數
// HSET Key hashKey jsonBody
// Key用server name
// hashKey則是將room name以City Hash32做hash後得出一個數字，以這個數字當hashKey
func (d *cache) AddServerOnline(server string, online *Online) (err error) {
	roomsMap := map[uint32]map[string]int32{}
	for room, count := range online.RoomCount {
		rMap := roomsMap[cityhash.CityHash32([]byte(room), uint32(len(room)))%8]
		if rMap == nil {
			rMap = make(map[string]int32)
			roomsMap[cityhash.CityHash32([]byte(room), uint32(len(room)))%8] = rMap
		}
		rMap[room] = count
	}

	key := keyServerOnline(server)
	for hashKey, value := range roomsMap {
		err = d.addServerOnline(key, strconv.FormatInt(int64(hashKey), 10), &Online{RoomCount: value, Server: online.Server, Updated: online.Updated})
		if err != nil {
			return
		}
	}
	return
}

// 以HSET方式儲存房間人數
// HSET Key hashKey jsonBody
// Key用server name
func (d *cache) addServerOnline(key string, hashKey string, online *Online) (err error) {
	conn := d.Get()
	defer conn.Close()
	b, _ := json.Marshal(online)
	if err = conn.Send("HSET", key, hashKey, b); err != nil {
		log.Errorf("conn.Send(SET %s,%s) error(%v)", key, hashKey, err)
		return
	}
	if err = conn.Send("EXPIRE", key, d.expire); err != nil {
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
func (d *cache) ServerOnline(server string) (online *Online, err error) {
	online = &Online{RoomCount: map[string]int32{}}
	// server name
	key := keyServerOnline(server)
	for i := 0; i < 8; i++ {
		ol, err := d.serverOnline(key, strconv.FormatInt(int64(i), 10))
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
func (d *cache) serverOnline(key string, hashKey string) (online *Online, err error) {
	conn := d.Get()
	defer conn.Close()
	// b是一個json
	// {
	// 		"server":"ne0002de-MacBook-Pro.local",
	// 		"room_count":{
	// 			"1000":1
	// 		 },
	// 		 "updated":1556368160
	// }"
	// 1000是房間id，1是人數
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
func (d *cache) DelServerOnline(server string) (err error) {
	conn := d.Get()
	defer conn.Close()
	key := keyServerOnline(server)
	if _, err = conn.Do("DEL", key); err != nil {
		log.Errorf("conn.Do(DEL %s) error(%v)", key, err)
	}
	return
}
