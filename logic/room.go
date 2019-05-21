package logic

import (
	"database/sql"
	log "github.com/golang/glog"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/store"
)

func (l *Logic) SetRoom(r store.Room) bool {
	if aff, err := l.db.SetRoom(r); err != nil || aff <= 0 {
		log.Errorf("l.db.SetRoom(room: %v) error(%v)", r, err)
		return false
	}
	return true
}

func (l *Logic) GetRoom(roomId int) (store.Room, bool) {
	r, err := l.db.GetRoom(roomId)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Errorf("l.db.GetRoom(roomId: %d) error(%v)", roomId, err)
		}
		return r, false
	}
	return r, true
}
