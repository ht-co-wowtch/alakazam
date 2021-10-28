package scheme

import (
	"encoding/json"
	"gitlab.com/jetfueltw/cpw/alakazam/app/comet/pb"
	logicpb "gitlab.com/jetfueltw/cpw/alakazam/app/logic/pb"
)

const (
	// Display 會員等級 背景色
	// Lv 1~10
	DisplayLevel1BackgroundColor = "#4BB679"
	// Lv 11~20
	DisplayLevel2BackgroundColor = "#009A57"
	// Lv 21~30
	DisplayLevel3BackgroundColor = "#0099E4"
	// Lv 31~40
	DisplayLevel4BackgroundColor = "#006EB9"
	// Lv 41~50
	DisplayLevel5BackgroundColor = "#A5A131"
	// Lv 51~60
	DisplayLevel6BackgroundColor = "#E1A709"
	// Lv 61~70
	DisplayLevel7BackgroundColor = "#F15D22"
	// Lv 71~80
	DisplayLevel8BackgroundColor = "#DB4B82"
	// Lv 81~90
	DisplayLevel9BackgroundColor = "#FF3569"
	// Lv 91~
	DisplayLevel10BackgroundColor = "#A30015"
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

func levelBackgroundColor(lv int) string {
	color := DisplayLevel1BackgroundColor
	switch level := lv; {
	case level < 11:
		color = DisplayLevel1BackgroundColor
	case level < 21:
		color = DisplayLevel2BackgroundColor
	case level < 31:
		color = DisplayLevel3BackgroundColor
	case level < 41:
		color = DisplayLevel4BackgroundColor
	case level < 51:
		color = DisplayLevel5BackgroundColor
	case level < 61:
		color = DisplayLevel6BackgroundColor
	case level < 71:
		color = DisplayLevel7BackgroundColor
	case level < 81:
		color = DisplayLevel8BackgroundColor
	case level < 91:
		color = DisplayLevel9BackgroundColor
	case level >= 91:
		color = DisplayLevel10BackgroundColor
	}

	return color
}
