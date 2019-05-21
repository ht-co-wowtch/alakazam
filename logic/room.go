package logic

import (
	"gitlab.com/jetfueltw/cpw/alakazam/logic/store"
	log "github.com/golang/glog"
)

func (l *Logic) SetRoom(r store.Room) bool {
	if aff, err := l.db.SetRoom(r); err != nil || aff <= 0 {
		log.Errorf("l.db.SetRoom(room: %v) error(%v)", r, err)
		return false
	}
	return true
}
