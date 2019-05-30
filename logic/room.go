package logic

import (
	"database/sql"
	"fmt"
	log "github.com/golang/glog"
	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/client"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/permission"
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
	if _, err := l.db.UpdateRoom(r); err != nil {
		log.Errorf("l.db.CreateRoom(room: %v) error(%v)", r, err)
		return false
	}
	if err := l.cache.SetRoom(id, permission.ToRoomInt(r), r.Limit.Day, r.Limit.Dml, r.Limit.Amount); err != nil {
		log.Errorf("Logic UpdateRoom cache SetRoom(id:%s) error(%v)", id, err)
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

func (l *Logic) GetRoomPermission(rId string) int {
	i, err := l.cache.GetRoom(rId)

	if err != nil && err != redis.ErrNil {
		log.Errorf("Logic isBanned cache GetRoom(id:%s) error(%v) ", rId, err)
	}
	if i == 0 {
		var day, dml, amount int
		room, err := l.db.GetRoom(rId)

		if err == nil {
			i = permission.ToRoomInt(room)
			day = room.Limit.Day
			dml = room.Limit.Dml
			amount = room.Limit.Amount
		} else {
			i = permission.RoomDefaultPermission
			if err != sql.ErrNoRows {
				log.Errorf("Logic isBanned db GetRoom(id:%s) error(%v) ", rId, err)
			}
		}

		if err := l.cache.SetRoom(rId, i, day, dml, amount); err != nil {
			log.Errorf("Logic isBanned cache SetRoom(id:%s) error(%v) ", rId, err)
		}
	}
	return i
}

func (l *Logic) isMessage(uid, rid string, status int) error {
	if !permission.IsMoney(status) {
		return nil
	}

	day, dml, amount, err := l.cache.GetRoomByMoney(rid)
	if err != nil {
		log.Errorf("Logic isMessage cache GetRoomByMoney(room id:%s) error(%v)", rid, err)
		return errors.FailureError
	}

	money, err := l.client.GetMoney(day, &client.Option{Uid: uid, Token: ""})
	if err != nil {
		log.Errorf("Logic isMessage client GetMoney(id:%s day:%d) error(%v)", uid, day, err)
		return errors.FailureError
	}

	if dml > money.Dml || amount > money.Deposit {
		return errors.MoneyError.Format(day, amount, dml)
	}
	return nil
}
