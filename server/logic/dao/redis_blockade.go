package dao


import (
	
	log "github.com/golang/glog"
	
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic/business"
	
	"math/big"
)

//SetBlockade ...
func (d *Dao) SetBlockade(uid string, remark string ) (err error) {
	sec := big.NewInt(1 <<62 )
	conn := d.redis.Get()
	defer conn.Close()
	if err := conn.Send("SET", keyBlockadeInfo(uid), sec, business.Blockade); err != nil {
		log.Errorf("conn.Send(SET %s) error(%v)", uid, err)
		return err
	}
	if err = conn.Send("HINCRBY", keyUidInfo(uid), hashStatusKey, -business.Blockade); err != nil {
		log.Errorf("conn.Send(HSET %s) error(%v)", uid, err)
		return
	}
	if err = conn.Send("EXPIRE", keyBannedInfo(uid), sec); err != nil {
		log.Errorf("conn.Send(EXPIRE %s) error(%v)", uid, err)
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