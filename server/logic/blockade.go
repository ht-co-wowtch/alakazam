package logic

import (
	log "github.com/golang/glog"
)

func (l *Logic) SetBlockade(uid, remark string) error {
	if err := l.dao.SetBlockade(uid, remark); err != nil {
		log.Errorf("logic.SetBlockade uid:%s error(%v)", uid, err)
		return err
	}
	return nil
}

func (l *Logic) RemoveBlockade(uid string) error {
	if err := l.dao.RemoveBlockade(uid); err != nil {
		log.Errorf("logic.RemoveBlockade uid:%s error(%v)", uid, err)
		return err
	}
	return nil
}
