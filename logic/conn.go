package logic

import (
	"encoding/json"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/cache"
	"time"
)

type ConnectReply struct {
	// user uid
	Uid string

	// websocket connection key
	Key string

	// user name
	Name string

	// key所在的房間id
	RoomId string

	// 前台心跳週期時間
	Hb int64

	// 操作權限
	Permission int
}

// redis紀錄某人連線資訊
func (l *Logic) Connect(server string, token []byte) (*ConnectReply, error) {
	var params struct {
		// 帳務中心+版的認證token
		Token string `json:"token"`

		// client要進入的room
		RoomID string `json:"room_id"`
	}
	if err := json.Unmarshal(token, &params); err != nil {
		return nil, err
	}

	r := new(ConnectReply)
	user, key, err := l.login(params.Token, params.RoomID, server)
	if err != nil {
		return nil, err
	}
	r.Uid = user.Uid
	r.Name = user.Name
	r.Permission = user.Status()
	r.RoomId = params.RoomID
	r.Key = key
	// 告知comet連線多久沒心跳就直接close
	r.Hb = l.c.Heartbeat
	return r, nil
}

// redis清除某人連線資訊
func (l *Logic) Disconnect(uid, key, server string) (has bool, err error) {
	return l.cache.DeleteUser(uid, key)
}

// user key更換房間
func (l *Logic) ChangeRoom(uid, key, roomId string) error {
	return l.cache.ChangeRoom(uid, key, roomId)
}

// 更新某人redis資訊的過期時間
func (l *Logic) Heartbeat(uid, key, roomId, name, server string) error {
	_, err := l.cache.RefreshUserExpire(uid)
	if err != nil {
		return err
	}
	return nil
}

// restart redis內存的每個房間總人數
func (l *Logic) RenewOnline(server string, roomCount map[string]int32) (map[string]int32, error) {
	online := &cache.Online{
		Server:    server,
		RoomCount: roomCount,
		Updated:   time.Now().Unix(),
	}
	if err := l.cache.AddServerOnline(server, online); err != nil {
		return nil, err
	}
	return l.roomCount, nil
}
