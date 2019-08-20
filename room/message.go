package room

import (
	"gitlab.com/jetfueltw/cpw/alakazam/app/logic/pb"
	"gitlab.com/jetfueltw/cpw/alakazam/message"
	"time"
)

func (r *Room) GetMessage(roomId string) ([]interface{}, error) {
	room, err := r.Get(roomId)
	if err != nil {
		return nil, err
	}
	msg, err := r.db.GetRoomMessage(room.Id)
	if err != nil {
		return nil, err
	}

	data := make([]interface{}, 0)
	for i, v := range msg.List {
		switch v {
		case pb.PushMsg_MONEY:
			data = append(data, message.Money{
				message.Message{
					Id:      i,
					Type:    pb.PushMsg_MONEY,
					Message: msg.RedEnvelopeMessage[i].Message,
					Time:    msg.RedEnvelopeMessage[i].SendAt.Format(time.RFC3339),
				},
				message.RedEnvelope{
					Id:      msg.RedEnvelopeMessage[i].RedEnvelopesId,
					Token:   msg.RedEnvelopeMessage[i].Token,
					Expired: msg.RedEnvelopeMessage[i].ExpireAt.Unix(),
				},
			})
		default:
			data = append(data, message.Message{
				Id:      i,
				Type:    pb.PushMsg_ROOM,
				Message: msg.Message[i].Message,
				Time:    msg.Message[i].SendAt.Format(time.RFC3339),
			})
		}
	}

	return data, nil
}
