package cache

import (
	"fmt"
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
// token => 三方應用接口token
// server => comet server name
func (d *Cache) SetUser(uid, key, roomId, name, token, server string, status int) (err error) {
	conn := d.Get()
	defer conn.Close()
	if err = conn.Send("HMSET", keyUidInfo(uid), key, roomId, hashNameKey, name, hashStatusKey, status, hashTokenKey, token, hashServerKey, server); err != nil {
		log.Errorf("conn.Send(HMSET %s,%s,%s,%s,%d,%s) error(%v)", uid, key, roomId, name, status, server, err)
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
func (d *Cache) RefreshUserExpire(uid string) (has bool, err error) {
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
func (d *Cache) DeleteUser(uid, key string) (has bool, err error) {
	conn := d.Get()
	defer conn.Close()
	if err = conn.Send("HDEL", keyUidInfo(uid), key); err != nil {
		log.Errorf("conn.Send(HDEL %s) error(%v)", uid, err)
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
	if err = conn.Send("HMGET", keyUidInfo(uid), key, hashNameKey, hashStatusKey); err != nil {
		log.Errorf("conn.Do(HMGET %s) error(%v)", uid, err)
		return "", "", 0, err
	}
	if err = conn.Flush(); err != nil {
		log.Errorf("conn.Flush() error(%v)", err)
		return "", "", 0, err
	}
	res, err := redis.Strings(conn.Receive())
	if err != nil {
		log.Errorf("conn.Receive() error(%v)", err)
		return "", "", 0, err
	}

	if len(res) != 3 {
		return "", "", 0, fmt.Errorf("conn.Receive() len is %d insufficient 3", len(res))
	}

	status, err = strconv.Atoi(res[2])

	return res[0], res[1], status, err
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
