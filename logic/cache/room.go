package cache

import (
	log "github.com/golang/glog"
	"github.com/gomodule/redigo/redis"
)

const (
	hashPermissionKey = "permission"

	hashLimitDayKey = "day"

	hashLimitAmountKey = "amount"

	hashLimitDmlKey = "dml"
)

var (
	roomExpired = 60 * 60
)

func (c *Cache) GetRoom(id string) (i int, err error) {
	conn := c.Get()
	defer conn.Close()
	if err = conn.Send("HGET", keyRoom(id), hashPermissionKey); err != nil {
		log.Errorf("GetRoom conn.Send(HGET %s) error(%v)", id, err)
		return 0, err
	}
	if err = conn.Flush(); err != nil {
		log.Errorf("GetRoom conn.Flush() error(%v)", err)
		return 0, err
	}
	return redis.Int(conn.Receive())
}

func (c *Cache) GetRoomByMoney(id string) (day, dml, amount int, err error) {
	conn := c.Get()
	defer conn.Close()
	if err = conn.Send("HMGET", keyRoom(id), hashLimitDayKey, hashLimitDmlKey, hashLimitAmountKey); err != nil {
		log.Errorf("GetRoomByMoney conn.Send(HMGET %s) error(%v)", id, err)
		return 0, 0, 0, err
	}
	if err = conn.Flush(); err != nil {
		log.Errorf("GetRoomByMoney conn.Flush() error(%v)", err)
		return 0, 0, 0, err
	}

	i, err := redis.Ints(conn.Receive())

	if err != nil {
		return 0, 0, 0, err
	}
	return i[0], i[1], i[2], err
}

func (c *Cache) SetRoom(id string, permission, day, dml, amount int) error {
	conn := c.Get()
	defer conn.Close()
	if err := conn.Send("HMSET", keyRoom(id), hashPermissionKey, permission, hashLimitDayKey, day, hashLimitDmlKey, dml, hashLimitAmountKey, amount); err != nil {
		log.Errorf("SetRoom conn.Send(HMSET key:%s permission:%d day:%d amount:%d dml:%d) error(%v)", id, permission, day, amount, dml, err)
		return err
	}
	if err := conn.Send("EXPIRE", keyRoom(id), roomExpired); err != nil {
		log.Errorf("SetRoom conn.Send(EXPIRE %s) error(%v)", id, err)
		return err
	}
	if err := conn.Flush(); err != nil {
		log.Errorf("SetRoom conn.Flush() error(%v)", err)
		return err
	}
	for i := 0; i < 2; i++ {
		if _, err := conn.Receive(); err != nil {
			log.Errorf("SetRoom conn.Receive() error(%v)", err)
			return err
		}
	}
	return nil
}
