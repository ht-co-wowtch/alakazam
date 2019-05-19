package dao

import (
	log "github.com/golang/glog"
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic/business"
)

// 設定封鎖
func (c *cache) SetBlockade(uid string) error {
	conn := c.Get()
	defer conn.Close()
	if err := conn.Send("HSET", keyUidInfo(uid), hashStatusKey, business.Blockade); err != nil {
		log.Errorf("conn.Send(HSET %s) error(%v)", uid, err)
		return err
	}
	if err := conn.Flush(); err != nil {
		log.Errorf("conn.Flush() error(%v)", err)
		return err
	}
	if _, err := conn.Receive(); err != nil {
		log.Errorf("conn.Receive() error(%v)", err)
		return err
	}
	return nil
}

// 解除封鎖
func (d *cache) RemoveBlockade(uid string) error {
	conn := d.Get()
	defer conn.Close()
	if err := conn.Send("HSET", keyUidInfo(uid), hashStatusKey, business.PlayDefaultPermission); err != nil {
		log.Errorf("conn.Send(HSET %s) error(%v)", uid, err)
		return err
	}
	if err := conn.Flush(); err != nil {
		log.Errorf("conn.Flush() error(%v)", err)
		return err
	}
	if _, err := conn.Receive(); err != nil {
		log.Errorf("conn.Receive() error(%v)", err)
		return err
	}
	return nil
}
