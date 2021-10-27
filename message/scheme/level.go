package scheme

import (
	"encoding/json"
	"gitlab.com/jetfueltw/cpw/alakazam/app/comet/pb"
	logicpb "gitlab.com/jetfueltw/cpw/alakazam/app/logic/pb"
)

type LevelUpAlertMessage struct {
	Message
	Level int `json:"level"`
}

func (m LevelUpAlertMessage) ToProto() (*logicpb.PushMsg, error) {
	bm, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return &logicpb.PushMsg{
		Seq:    m.Id,
		Op:     pb.OpRaw,
		Msg:    bm,
		SendAt: m.Timestamp,
	}, nil
}