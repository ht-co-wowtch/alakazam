package logic

import (
	"github.com/go-redis/redis"
	log "github.com/golang/glog"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
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
	Deposit int `json:"deposit"`

	// 打碼量
	Dml int `json:"dml"`
}

func (l *Logic) CreateRoom(r Room) (string, error) {
	room := models.Room{
		Id:                r.Id,
		IsMessage:         r.IsMessage,
		IsFollow:          r.IsFollow,
		DayLimit:          r.Limit.Day,
		DepositLimit:      r.Limit.Deposit,
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
	return r.Id, l.cache.SetRoom(room)
}

func (l *Logic) UpdateRoom(r Room) bool {
	room := models.Room{
		Id:           r.Id,
		IsMessage:    r.IsMessage,
		IsFollow:     r.IsFollow,
		DayLimit:     r.Limit.Day,
		DepositLimit: r.Limit.Deposit,
		DmlLimit:     r.Limit.Dml,
	}
	if _, err := l.db.UpdateRoom(room); err != nil {
		log.Errorf("l.db.CreateRoom(room: %v) error(%v)", r, err)
		return false
	}
	if err := l.cache.SetRoom(room); err != nil {
		log.Errorf("Logic UpdateRoom cache SetRoom(id:%s) error(%v)", r.Id, err)
		return false
	}
	return true
}

func (l *Logic) GetRoom(roomId string) (models.Room, bool) {
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
		room, _, err := l.db.GetRoom(rId)
		// TODO 需要error判斷回傳值
		if err != nil {
			return 0
		}
		i = room.Status()
		if err := l.cache.SetRoom(room); err != nil {
			log.Errorf("Logic isBanned cache SetRoom(id:%s) error(%v) ", rId, err)
		}
	}
	return i
}

func (l *Logic) isMessage(rid string, status int, uid, token string) error {
	if !models.IsMoney(status) {
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
