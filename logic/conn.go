package logic

import (
	"encoding/json"
	"fmt"
	log "github.com/golang/glog"
	"github.com/google/uuid"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/cache"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/permission"
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
		log.Errorf("json.Unmarshal(%s) error(%v)", token, err)
		return nil, errors.ConnectError
	}

	r := new(ConnectReply)

	user, err := l.auth(params.Token)
	if err != nil {
		return r, err
	}

	// 封鎖會員
	if user.IsBlockade {
		r.Permission = permission.Blockade
		return r, nil
	}

	r.Uid = user.Uid
	r.Name = user.Name
	r.Permission = user.Permission
	r.RoomId = params.RoomID

	// 告知comet連線多久沒心跳就直接close
	r.Hb = l.c.Heartbeat

	r.Key = uuid.New().String()

	// 儲存user資料至redis
	if err := l.cache.SetUser(r.Uid, r.Key, r.RoomId, r.Name, "test", server, r.Permission); err != nil {
		log.Errorf("l.dao.SetUser(%s,%s,%s,%s) error(%v)", r.Uid, r.Key, r.Name, server, err)
	}
	log.Infof("conn connected key:%s server:%s uid:%s token:%s", r.Key, server, r.Uid, token)
	return r, nil
}

// redis清除某人連線資訊
func (l *Logic) Disconnect(uid, key, server string) (has bool, err error) {
	if has, err = l.cache.DeleteUser(uid, key); err != nil {
		log.Errorf("l.dao.DeleteUser(%s,%s) error(%v)", uid, key, err)
		return
	}
	log.Infof("conn disconnected server:%s uid:%s key:%s", server, uid, key)
	return
}

// user key更換房間
func (l *Logic) ChangeRoom(uid, key, roomId string) (err error) {
	if err = l.cache.ChangeRoom(uid, key, roomId); err != nil {
		log.Errorf("l.dao.DeleteUser(%s,%s)", uid, key)
		return
	}
	log.Infof("conn ChangeRoom  uid:%s key:%s roomId:%s", uid, key, roomId)
	return
}

// 更新某人redis資訊的過期時間
func (l *Logic) Heartbeat(uid, key, roomId, name, server string) error {
	has, err := l.cache.RefreshUserExpire(uid)
	if err != nil {
		log.Errorf("l.dao.RefreshUserExpire(%s,%s,%s) error(%v)", uid, key, server, err)
		return err
	}
	// 沒更新成功就直接做覆蓋
	if !has {
		e := fmt.Errorf("Heartbeat(uid:%s key:%s server:%s) error(%v)", uid, key, server, err)
		log.Error(e)
		return e
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
