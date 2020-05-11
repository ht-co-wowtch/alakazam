package http

import (
	"fmt"
	"gitlab.com/jetfueltw/cpw/alakazam/client"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/member"
	"gitlab.com/jetfueltw/cpw/alakazam/message"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"gitlab.com/jetfueltw/cpw/alakazam/room"
	"gitlab.com/jetfueltw/cpw/micro/errdefs"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"time"
)

type msg struct {
	room    room.Chat
	client  *client.Client
	message *message.Producer
	member  *member.Member
}

func (m *msg) user(req messageReq) (int64, error) {
	user, chat, err := m.room.GetUserMessageSession(req.uid, req.RoomId)
	if err != nil {
		return 0, err
	}

	if chat.DayLimit >= 1 && chat.DmlLimit+chat.DepositLimit > 0 {
		money, err := m.client.GetDepositAndDml(chat.DayLimit, user.Uid, req.token)
		if err != nil {
			return 0, err
		}
		if float64(chat.DmlLimit) > money.Dml || float64(chat.DepositLimit) > money.Deposit {
			msg := fmt.Sprintf(errors.ErrRoomLimit, chat.DayLimit, chat.DepositLimit, chat.DmlLimit)
			return 0, errdefs.Unauthorized(5005, msg)
		}
	}

	var display message.Display
	u := toUserMessage(user)
	if user.Type == models.STREAMER {
		display = message.DisplayByUser(u, req.Message)
	} else {
		display = message.DisplayByStreamer(u, req.Message)
	}

	msg := message.ProducerMessage{
		Rooms:   []int32{int32(req.RoomId)},
		User:    u,
		Display: display,
	}

	id, err := m.message.Send(msg)

	if err == errors.ErrRateSameMsg {
		isBlockade, err := m.member.SetBannedForSystem(user.Uid, 10*60)
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

	u := toUserMessage(user)
	msg := message.ProducerMessage{
		Rooms:   []int32{int32(req.RoomId)},
		User:    u,
		Display: message.DisplayByUser(u, req.Message),
	}

	redEnvelope := message.RedEnvelope{
		Id:      reply.Order,
		Token:   reply.Token,
		Expired: reply.ExpireAt.Format(time.RFC3339),
	}

	msgId, err := m.message.SendRedEnvelope(msg, redEnvelope)

	if err != nil {
		log.Error("send red envelope message error",
			zap.Error(err),
			zap.String("uid", user.Uid),
			zap.String("order", reply.Order),
		)
	}

	return msgId, reply, err
}

func toUserMessage(user *models.Member) message.User {
	return message.User{
		Id:     int64(user.Id),
		Uid:    user.Uid,
		Name:   user.Name,
		Avatar: message.ToAvatarName(user.Gender),
	}
}
