package scheme

import (
	"encoding/json"
	"gitlab.com/ht-co/wowtch/live/alakazam/app/comet/pb"
	logicpb "gitlab.com/ht-co/wowtch/live/alakazam/app/logic/pb"
	"time"
)

type Follow struct {
	Message
	Follow interface{} `json:"follow"`
}

func NewFollowProto(seq int64, rid int32, user User, total int) (*logicpb.PushMsg, error) {
	now := time.Now()

	m := Follow{
		Message: Message{
			Id:        seq,
			Type:      FOLLOW,
			User:      user,
			Time:      now.Format("15:04:05"),
			Timestamp: now.Unix(),
		},
		Follow: struct {
			Total int `json:"total"`
		}{Total: total},
	}

	bm, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	return &logicpb.PushMsg{
		Room:   []int32{rid},
		Seq:    seq,
		Type:   logicpb.PushMsg_ROOM,
		Op:     pb.OpRaw,
		Msg:    bm,
		SendAt: m.Message.Timestamp,
	}, nil
}
