package http

import (
	"fmt"
	"time"

	"gitlab.com/ht-co/wowtch/live/alakazam/client"
	"gitlab.com/ht-co/wowtch/live/alakazam/errors"
	"gitlab.com/ht-co/wowtch/live/alakazam/member"
	"gitlab.com/ht-co/wowtch/live/alakazam/message"
	"gitlab.com/ht-co/wowtch/live/alakazam/message/scheme"
	"gitlab.com/ht-co/wowtch/live/alakazam/room"
	"gitlab.com/ht-co/cpw/micro/errdefs"
	"gitlab.com/ht-co/cpw/micro/log"
	"go.uber.org/zap"
)

type msg struct {
	room    room.Chat
	client  *client.Client
	message *message.Producer
	member  *member.Member
}

// 發訊息給房間所有人
func (m *msg) user(req messageReq) (int64, error) {
	user, chat, err := m.room.GetMessageSession(req.Uid, req.RoomId)
	if err != nil {
		return 0, err
	}

	if chat.DayLimit >= 1 && chat.DmlLimit+chat.DepositLimit > 0 {
		money, err := m.client.GetDepositAndDml(chat.DayLimit, user.Uid, req.Token)
		if err != nil {
			return 0, err
		}
		if float64(chat.DmlLimit) > money.Dml || float64(chat.DepositLimit) > money.Deposit {
			msg := fmt.Sprintf(errors.ErrRoomLimit, chat.DayLimit, chat.DepositLimit, chat.DmlLimit)
			return 0, errdefs.Unauthorized(5005, msg)
		}
	}

	id, err := m.message.SendUser([]int32{int32(req.RoomId)}, req.Message, user)

	if err == errors.ErrRateSameMsg {
		isBlockade, err := m.member.SetBannedForSystem(user.Uid, req.RoomId, 10*60)
		if err != nil {
			log.Error("set banned for rate same message", zap.Error(err), zap.String("uid", user.Uid))
		}
		if isBlockade {
			keys, err := m.member.Kick(user.Uid)
			if err != nil {
				log.Error("kick member for push room", zap.Error(err), zap.String("uid", user.Uid))
			}
			if len(keys) > 0 {
				err = m.message.Kick("你被踢出房间，因为自动禁言达五次", keys)
				if err == nil {
					log.Error("kick member set message for push room", zap.Error(err))
				}
			}
		}
	}

	return id, err
}

// 發送訊息給房間內特定人
func (m *msg) private(req messageReq) (int64, error) {
	user, err := m.member.GetSession(req.Uid)
	if err != nil {
		return 0, err
	}

	keys, err := m.member.GetRoomKeys(req.ToUid, req.RoomId)
	if err != nil {
		return 0, err
	}

	if len(keys) == 0 {
		return 0, errors.ErrNoOnline
	}

	_, err = m.message.SendPrivate(keys, req.Message, user)

	// 主播端也需要接收到此訊息
	mKeys, err := m.member.GetKeys(req.Uid)
	if err != nil {
		return 0, err
	}

	user, err = m.member.GetSession(req.ToUid)
	if err != nil {
		return 0, err
	}

	return m.message.SendPrivateReply(mKeys, user)
}

func (m *msg) redEnvelope(req giveRedEnvelopeReq) (int64, client.RedEnvelopeReply, error) {
	user, reply, err := m.member.GiveRedEnvelope(req.uid, req.token, member.RedEnvelope{
		RoomId:  req.RoomId,
		Message: req.Message,
		Type:    req.Type,
		Amount:  req.Amount,
		Count:   req.Count,
	})
	if err != nil {
		return 0, client.RedEnvelopeReply{}, err
	}

	redEnvelope := scheme.RedEnvelope{
		Id:      reply.Order,
		Token:   reply.Token,
		Expired: reply.ExpireAt.Format(time.RFC3339),
	}

	msgId, err := m.message.SendRedEnvelope([]int32{int32(req.RoomId)}, req.Message, scheme.NewUser(*user), redEnvelope)

	if err != nil {
		log.Error("[message.go]redEnvelope",
			zap.Error(err),
			zap.String("uid", user.Uid),
			zap.String("order", reply.Order),
		)
	}

	return msgId, reply, err
}
