package cache

import (
	log "github.com/golang/glog"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/permission"
	"strconv"
	"time"
)

// 設定禁言
func (d *Cache) SetBanned(uid string, expired int) (err error) {
	sec := time.Duration(expired) * time.Second
	conn := d.Get()
	defer conn.Close()
	if err := conn.Send("SET", keyBannedInfo(uid), time.Now().Add(sec).Unix()); err != nil {
		log.Errorf("conn.Send(SET %s) error(%v)", uid, err)
		return err
	}
	if err = conn.Send("EXPIRE", keyBannedInfo(uid), expired); err != nil {
		log.Errorf("conn.Send(EXPIRE %s) error(%v)", uid, err)
		return
	}
	if err = conn.Send("HINCRBY", keyUidInfo(uid), hashStatusKey, -permission.Message); err != nil {
		log.Errorf("conn.Send(HSET %s) error(%v)", uid, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Errorf("conn.Flush() error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Errorf("conn.Receive() error(%v)", err)
			return err
		}
	}
	return nil
}

// 取得禁言時效
func (d *Cache) GetBanned(uid string) (t time.Time, is bool, err error) {
	conn := d.Get()
	defer conn.Close()
	var r interface{}
	if r, err = conn.Do("GET", keyBannedInfo(uid)); err != nil {
		log.Errorf("conn.Send(GET %s) error(%v)", uid, err)
		return
	}

	switch reply := r.(type) {
	case int64:
		t = time.Unix(int64(reply), 0)
		is = true
		return
	case nil:
		return
	case []byte:
		n, e := strconv.ParseInt(string(reply), 10, 64)
		t = time.Unix(n, 0)
		is = true
		err = e
		return
	}
	return
}

// 解除禁言
func (d *Cache) DelBanned(uid string) (err error) {
	conn := d.Get()
	defer conn.Close()
	if err = conn.Send("DEL", keyBannedInfo(uid)); err != nil {
		log.Errorf("conn.Send(DEL %s) error(%v)", uid, err)
		return
	}
	if err = conn.Send("HINCRBY", keyUidInfo(uid), hashStatusKey, permission.Message); err != nil {
		log.Errorf("conn.Send(HINCRBY %s) error(%v)", uid, err)
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
