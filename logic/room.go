package logic

import (
	"database/sql"
	"fmt"
	log "github.com/golang/glog"
	"github.com/google/uuid"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/store"
)

func (l *Logic) CreateRoom(r store.Room) (string, error) {
	id, _ := uuid.New().MarshalBinary()
	r.RoomId = fmt.Sprintf("%x", id)

	if aff, err := l.db.CreateRoom(r); err != nil || aff <= 0 {
		log.Errorf("l.db.CreateRoom(room: %v) error(%v)", r, err)
		return "", err
	}
	return r.RoomId, nil
}

func (l *Logic) UpdateRoom(id string, r store.Room) bool {
	r.RoomId = id
	if aff, err := l.db.UpdateRoom(r); err != nil || aff <= 0 {
		log.Errorf("l.db.CreateRoom(room: %v) error(%v)", r, err)
		return false
	}
	return true
}

func (l *Logic) GetRoom(roomId string) (store.Room, bool) {
	r, err := l.db.GetRoom(roomId)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Errorf("l.db.GetRoom(roomId: %s) error(%v)", roomId, err)
		}
		return r, false
	}
	return r, true
}
