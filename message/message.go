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

const (
	// 普通訊息
	messageType = "message"
	// 紅包訊息
	redEnvelopeType = "red_envelope"
	// 公告訊息
	topType = "top"
)

func (h *History) Get(roomId, lastMsgId int) ([]interface{}, error) {
	msg, err := h.db.GetRoomMessage(roomId, lastMsgId)
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
	for _, msgId := range msg.List {
		switch msg.Type[msgId] {
		case pb.PushMsg_MONEY:
			data = append(data, historyRedEnvelopeMessage{
				historyMessage: historyMessage{
					Id:      msgId,
					Uid:     memberMap[msg.RedEnvelopeMessage[msgId].MemberId].Uid,
					Name:    memberMap[msg.RedEnvelopeMessage[msgId].MemberId].Name,
					Type:    redEnvelopeType,
					Message: msg.RedEnvelopeMessage[msgId].Message,
					Time:    msg.RedEnvelopeMessage[msgId].SendAt.Format("15:04:05"),
				},
				RedEnvelope: historyRedEnvelope{
					Id:      msg.RedEnvelopeMessage[msgId].RedEnvelopesId,
					Token:   msg.RedEnvelopeMessage[msgId].Token,
					Expired: msg.RedEnvelopeMessage[msgId].ExpireAt.Format(time.RFC3339),
				},
			})
		case pb.PushMsg_ROOM:
			data = append(data, historyMessage{
				Id:      msgId,
				Uid:     memberMap[msg.Message[msgId].MemberId].Uid,
				Name:    memberMap[msg.Message[msgId].MemberId].Name,
				Type:    messageType,
				Message: msg.Message[msgId].Message,
				Time:    msg.Message[msgId].SendAt.Format("15:04:05"),
			})
		}
	}
	return data, nil
}
