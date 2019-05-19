package dao

import (
	log "github.com/golang/glog"
	"github.com/gomodule/redigo/redis"
	"strconv"
)

// 儲存user資訊
// HSET :
// 主key => uid_{user id}
// user key => user roomId
// name => user name
// status => user status
// server => comet server name
func (d *Cache) AddMapping(uid, key, roomId, name, server string, status int) (err error) {
	conn := d.Get()
	defer conn.Close()
	if err = conn.Send("HSET", keyUidInfo(uid), key, roomId, hashNameKey, name, hashStatusKey, status, hashServerKey, server); err != nil {
		log.Errorf("conn.Send(HSET %s,%s) error(%v)", uid, key, err)
		return
	}
	if err = conn.Send("EXPIRE", keyUidInfo(uid), d.expire); err != nil {
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
func (d *Cache) ExpireMapping(uid string) (has bool, err error) {
	conn := d.Get()
	defer conn.Close()
	if err = conn.Send("EXPIRE", keyUidInfo(uid), d.expire); err != nil {
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
func (d *Cache) DelMapping(uid, key, server string) (has bool, err error) {
	conn := d.Get()
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

func (d *Cache) GetUser(uid string, key string) (roomId, name string, status int, err error) {
	conn := d.Get()
	defer conn.Close()
	if err = conn.Send("HGETALL", keyUidInfo(uid)); err != nil {
		log.Errorf("conn.Do(HGETALL %s) error(%v)", uid, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Errorf("conn.Flush() error(%v)", err)
		return
	}
	var res map[string]string
	res, err = redis.StringMap(conn.Receive())
	if err != nil {
		log.Errorf("conn.Receive() error(%v)", err)
		return
	}

	// TODO 自行實作redis.StringMap
	status, err = strconv.Atoi(res[hashStatusKey])

	return res[key], res[hashNameKey], status, err
}

// 更換房間
func (d *Cache) ChangeRoom(uid, key, roomId string) (err error) {
	conn := d.Get()
	defer conn.Close()
	if err = conn.Send("HSET", keyUidInfo(uid), key, roomId); err != nil {
		log.Errorf("conn.Send(HSET %s,%s) error(%v)", uid, key, err)
		return
	}
	if err = conn.Send("EXPIRE", keyUidInfo(uid), d.expire); err != nil {
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