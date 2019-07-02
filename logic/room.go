package logic

import (
	log "github.com/golang/glog"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"gitlab.com/jetfueltw/cpw/micro/id"
)

type Room struct {
	// 要設定的房間id
	Id string `json:"id" binding:"len=32"`

	// 是否禁言
	IsMessage bool `json:"is_message"`

	// 是否可發/跟注
	IsFollow bool `json:"is_follow"`

	// 儲值&打碼量發話限制
	Limit Limit `json:"limit"`

	// 紅包多久過期
	RedEnvelopeExpire int `json:"red_envelope_expire" binding:"required,max=120"`
}

type Limit struct {
	// 限制範圍
	Day int `json:"day" binding:"max=31"`

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

func (l *Logic) UpdateRoom(r Room) error {
	room := models.Room{
		Id:                r.Id,
		IsMessage:         r.IsMessage,
		IsFollow:          r.IsFollow,
		DayLimit:          r.Limit.Day,
		DepositLimit:      r.Limit.Deposit,
		DmlLimit:          r.Limit.Dml,
		RedEnvelopeExpire: r.RedEnvelopeExpire,
	}
	if _, err := l.db.UpdateRoom(room); err != nil {
		return err
	}
	if err := l.cache.SetRoom(room); err != nil {
		return err
	}
	return nil
}

func (l *Logic) GetRoom(roomId string) (models.Room, bool, error) {
	return l.db.GetRoom(roomId)
}

func (l *Logic) isMessage(rid string, status int, uid, token string) error {
	if !models.IsMoney(status) {
		return nil
	}

	day, dml, amount, err := l.cache.GetRoomByMoney(rid)
	if err != nil {
		log.Errorf("Logic isMessage cache GetRoomByMoney(room id:%s) error(%v)", rid, err)
		return err
	}

	money, err := l.client.GetDepositAndDml(day, uid, token)
	if err != nil {
		log.Errorf("Logic isMessage client GetDepositAndDml(id:%s day:%d) error(%v)", uid, day, err)
		return err
	}

	if dml > money.Dml || amount > money.Deposit {
		return errors.MoneyError.Format(day, amount, dml)
	}
	return nil
}
