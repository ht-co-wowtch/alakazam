package logic

import (
	"github.com/go-redis/redis"
	log "github.com/golang/glog"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/permission"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/store"
	"gitlab.com/jetfueltw/cpw/micro/id"
)

type Room struct {
	// 要設定的房間id
	Id string `json:"id"`

	// 是否禁言
	IsMessage bool `json:"is_message"`

	// 是否可發/跟注
	IsFollow bool `json:"is_follow"`

	// 儲值&打碼量發話限制
	Limit Limit `json:"limit"`

	// 紅包多久過期
	RedEnvelopeExpire int `json:"red_envelope_expire"`
}

type Limit struct {
	// 限制範圍
	Day int `json:"day"`

	// 儲值金額
	Amount int `json:"amount"`

	// 打碼量
	Dml int `json:"dml"`
}

func (l *Logic) CreateRoom(r Room) (string, error) {
	room := store.Room{
		Id:                r.Id,
		IsMessage:         r.IsMessage,
		IsFollow:          r.IsFollow,
		DayLimit:          r.Limit.Day,
		DepositLimit:      r.Limit.Amount,
		DmlLimit:          r.Limit.Dml,
		RedEnvelopeExpire: r.RedEnvelopeExpire,
	}
	if r.Id == "" {
		room.Id = id.UUid32()
	}
	if len(room.Id) != 32 {
		return "", errors.DataError
	}
	if aff, err := l.db.CreateRoom(room); err != nil || aff <= 0 {
		return "", err
	}
	return r.Id, nil
}

func (l *Logic) UpdateRoom(id string, r Room) bool {
	room := store.Room{
		Id:           id,
		IsMessage:    r.IsMessage,
		IsFollow:     r.IsFollow,
		DayLimit:     r.Limit.Day,
		DepositLimit: r.Limit.Amount,
		DmlLimit:     r.Limit.Dml,
	}
	if _, err := l.db.UpdateRoom(room); err != nil {
		log.Errorf("l.db.CreateRoom(room: %v) error(%v)", r, err)
		return false
	}
	if err := l.cache.SetRoom(id, permission.ToRoomInt(room), room.DayLimit, room.DmlLimit, room.DepositLimit); err != nil {
		log.Errorf("Logic UpdateRoom cache SetRoom(id:%s) error(%v)", id, err)
		return false
	}
	return true
}

func (l *Logic) GetRoom(roomId string) (store.Room, bool) {
	r, ok, err := l.db.GetRoom(roomId)
	if err != nil {
		return r, false
	}
	return r, ok
}

func (l *Logic) GetRoomPermission(rId string) int {
	i, err := l.cache.GetRoom(rId)

	if err != nil && err != redis.Nil {
		log.Errorf("Logic isBanned cache GetRoom(id:%s) error(%v) ", rId, err)
	}
	if i == 0 {
		var day, dml, amount int
		room, ok, err := l.db.GetRoom(rId)
		// TODO 需要error判斷回傳值
		if err != nil {
			return 0
		}
		if ok {
			i = permission.ToRoomInt(room)
			day = room.DayLimit
			dml = room.DmlLimit
			amount = room.DepositLimit
		} else {
			i = permission.RoomDefaultPermission
		}
		if err := l.cache.SetRoom(rId, i, day, dml, amount); err != nil {
			log.Errorf("Logic isBanned cache SetRoom(id:%s) error(%v) ", rId, err)
		}
	}
	return i
}

func (l *Logic) isMessage(rid string, status int, uid, token string) error {
	if !permission.IsMoney(status) {
		return nil
	}

	day, dml, amount, err := l.cache.GetRoomByMoney(rid)
	if err != nil {
		log.Errorf("Logic isMessage cache GetRoomByMoney(room id:%s) error(%v)", rid, err)
		return errors.FailureError
	}

	money, err := l.client.GetDepositAndDml(day, uid, token)
	if err != nil {
		log.Errorf("Logic isMessage client GetDepositAndDml(id:%s day:%d) error(%v)", uid, day, err)
		return errors.FailureError
	}

	if dml > money.Dml || amount > money.Deposit {
		return errors.MoneyError.Format(day, amount, dml)
	}
	return nil
}
