package logic

import (
	log "github.com/golang/glog"
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic/business"
)

func (l *Logic) SetBlockade(uid, remark string) error {
	// TODO 待實作 從redis hash table 找出status並改成封鎖狀態
	// business.Blockade
	return l.dao.SetBlockade(uid, remark)
}

func (l *Logic) RemoveBlockade(uid string) error {
	// TODO 待實作 從redis hash table 找出status並寫回原來status
	// 原來的status要從DB拿，先暫時預設用business.PlayDefaultPermission
	if err := l.dao.RemoveBlockade(uid, business.PlayDefaultPermission) ; err != nil {
		log.Errorf("dao.DelBanned(uid: %s) error(%v)", uid, err)
		return err
	}
	return nil
}
