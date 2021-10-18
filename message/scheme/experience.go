package scheme

import (
	logicpb "gitlab.com/jetfueltw/cpw/alakazam/app/logic/pb"
	"time"
)

// 建立取得等級提升提示Proto
func NewLevelUpAlertProto(seq int64, rid []int32, user User) (*logicpb.PushMsg, error) {
	now := time.Now()

	m :=Message{
		Id:        seq,
		Type:      LEVEL_TYPE,
		User:      user,
		Time:      now.Format("15:04:05"),
		Timestamp: now.Unix(),
	}
	return m.ToRoomProto(rid)
}

// 建立取得等級提升訊息Proto
func NewLevelUpProto(seq int64, rid []int32, user User, level int) (*logicpb.PushMsg, error) {
	now := time.Now()
	m := Message{
		Id:        seq,
		Type:      MESSAGE_TYPE,
		User:      user,
		Display:   displayByLevelUp(user, level), // TODO
		Time:      now.Format("15:04:05"),
		Timestamp: now.Unix(),
	}
	return m.ToRoomProto(rid)
}
