package scheme

import (
	"encoding/json"
	"gitlab.com/jetfueltw/cpw/alakazam/app/comet/pb"
	logicpb "gitlab.com/jetfueltw/cpw/alakazam/app/logic/pb"
)

const (
	// Display 會員等級 背景色
	// Lv 1~10
	DisplayLevel1BackgroundColor = "#43636B"
	// Lv 11~20
	DisplayLevel2BackgroundColor = "#255584"
	// Lv 21~30
	DisplayLevel3BackgroundColor = "#027EAF"
	// Lv 31~40
	DisplayLevel4BackgroundColor = "#197C62"
	// Lv 41~50
	DisplayLevel5BackgroundColor = "#56823F"
	// Lv 51~60
	DisplayLevel6BackgroundColor = "#87802E"
	// Lv 61~70
	DisplayLevel7BackgroundColor = "#CC8615"
	// Lv 71~80
	DisplayLevel8BackgroundColor = "#F15D22"
	// Lv 81~90
	DisplayLevel9BackgroundColor = "#FF52B9"
	// Lv 91~
	DisplayLevel10BackgroundColor = "#EF225D"

	// Lv 100
	DisplayLevel11BackgroundColor = "#9D1E0E"

	// 等級文字顏色
	DisplayLevelTextColor = "#FFFFFF"

	// 等級100文字顏色
	DisplayLevelTopTextColor = "#FFFF00"
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

// 等級區塊背景色
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
	case level < 100:
		color = DisplayLevel10BackgroundColor
	case level >= 100:
		color = DisplayLevel11BackgroundColor
	}

	return color
}

// 等級文字顏色
func levelTextColor(lv int) string {
	color := DisplayLevelTextColor

	if lv >= 100 {
		color = DisplayLevelTopTextColor
	}

	return color
}
