package cache

import (
	log "github.com/golang/glog"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/permission"
)

// 設定封鎖
func (c *Cache) SetBlockade(uid string) error {
	conn := c.Get()
	defer conn.Close()
	if err := conn.Send("HSET", keyUidInfo(uid), hashStatusKey, permission.Blockade); err != nil {
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
func (d *Cache) RemoveBlockade(uid string) error {
	conn := d.Get()
	defer conn.Close()
	if err := conn.Send("HSET", keyUidInfo(uid), hashStatusKey, permission.PlayDefaultPermission); err != nil {
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
