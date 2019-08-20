package message

import (
	"gitlab.com/jetfueltw/cpw/alakazam/app/logic/pb"
	"gitlab.com/jetfueltw/cpw/alakazam/member"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"time"
)

type History struct {
	db     *models.Store
	member *member.Member
}

func NewHistory(db *models.Store, member *member.Member) *History {
	return &History{
		db:     db,
		member: member,
	}
}

func (h *History) Get(roomId int) ([]interface{}, error) {
	msg, err := h.db.GetRoomMessage(roomId)
	if err != nil {
		return nil, err
	}

	mids := make([]int, len(msg.Message)+len(msg.RedEnvelopeMessage))

	for _, v := range msg.Message {
		mids = append(mids, v.MemberId)
	}
	for _, v := range msg.RedEnvelopeMessage {
		mids = append(mids, v.MemberId)
	}

	ms, err := h.member.GetMembers(mids)
	if err != nil {
		return nil, err
	}

	memberMap := make(map[int]models.Member, len(ms))
	for _, v := range ms {
		memberMap[v.Id] = v
	}

	data := make([]interface{}, 0)
	for i, v := range msg.List {
		switch v {
		case pb.PushMsg_MONEY:
			data = append(data, Money{
				Message{
					Id:      i,
					Uid:     memberMap[msg.RedEnvelopeMessage[i].MemberId].Uid,
					Name:    memberMap[msg.RedEnvelopeMessage[i].MemberId].Name,
					Type:    pb.PushMsg_MONEY,
					Message: msg.RedEnvelopeMessage[i].Message,
					Time:    msg.RedEnvelopeMessage[i].SendAt.Format(time.RFC3339),
				},
				RedEnvelope{
					Id:      msg.RedEnvelopeMessage[i].RedEnvelopesId,
					Token:   msg.RedEnvelopeMessage[i].Token,
					Expired: msg.RedEnvelopeMessage[i].ExpireAt.Unix(),
				},
			})
		default:
			data = append(data, Message{
				Id:      i,
				Uid:     memberMap[msg.Message[i].MemberId].Uid,
				Name:    memberMap[msg.Message[i].MemberId].Name,
				Type:    pb.PushMsg_ROOM,
				Message: msg.Message[i].Message,
				Time:    msg.Message[i].SendAt.Format(time.RFC3339),
			})
		}
	}

	return data, nil
}
