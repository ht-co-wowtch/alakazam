package cache

import (
	log "github.com/golang/glog"
	"github.com/gomodule/redigo/redis"
)

var (
	roomExpired = 60 * 60
)

func (c *Cache) GetRoom(id string) (i int, err error) {
	conn := c.Get()
	defer conn.Close()
	if err = conn.Send("GET", keyRoom(id)); err != nil {
		log.Errorf("GetRoom conn.Send(GET %s) error(%v)", id, err)
		return 0, err
	}
	if err = conn.Flush(); err != nil {
		log.Errorf("GetRoom conn.Flush() error(%v)", err)
		return 0, err
	}
	return redis.Int(conn.Receive())
}

func (c *Cache) SetRoom(id string, permission int) error {
	conn := c.Get()
	defer conn.Close()
	if err := conn.Send("SET", keyRoom(id), permission); err != nil {
		log.Errorf("SetRoom conn.Send(SET key:%s permission:%d) error(%v)", id, permission, err)
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
