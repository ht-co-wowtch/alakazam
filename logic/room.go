package logic

import (
	"fmt"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"gitlab.com/jetfueltw/cpw/micro/errdefs"
)

type Room struct {
	// 要設定的房間id
	Id string `json:"id" binding:"required,len=32"`

	// 是否禁言
	IsMessage bool `json:"is_message"`

	// 儲值&打碼量發話限制
	Limit Limit `json:"limit"`
}

type Limit struct {
	// 限制範圍
	Day int `json:"day" binding:"max=31"`

	// 儲值金額
	Deposit float64 `json:"deposit"`

	// 打碼量
	Dml float64 `json:"dml"`
}

func (l *Logic) CreateRoom(r Room) (string, error) {
	room := models.Room{
		Id:           r.Id,
		IsMessage:    r.IsMessage,
		DayLimit:     r.Limit.Day,
		DepositLimit: r.Limit.Deposit,
		DmlLimit:     r.Limit.Dml,
	}
	if _, err := l.db.CreateRoom(room); err != nil {
		return "", err
	}
	return r.Id, l.cache.SetRoom(room)
}

func (l *Logic) UpdateRoom(r Room) error {
	room := models.Room{
		Id:           r.Id,
		IsMessage:    r.IsMessage,
		DayLimit:     r.Limit.Day,
		DepositLimit: r.Limit.Deposit,
		DmlLimit:     r.Limit.Dml,
	}
	aff, err := l.db.UpdateRoom(room)
	if err != nil {
		return err
	}
	if aff <= 0 {
		return errors.ErrNoRows
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
	day, dml, deposit, err := l.cache.GetRoomByMoney(rid)
	if err != nil {
		return err
	}
	money, err := l.client.GetDepositAndDml(day, uid, token)
	if err != nil {
		return err
	}
	if float64(dml) > money.Dml || float64(deposit) > money.Deposit {
		e := errors.New(fmt.Sprintf("您无法发言，当前发言条件：前%d天充值不少于%d元；打码量不少于%d元", day, deposit, dml))
		return errdefs.Unauthorized(e, 4)
	}
	return nil
}
