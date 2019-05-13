package logic

import (
	"context"
	"encoding/json"
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic/dao"
	"time"

	log "github.com/golang/glog"
	"github.com/google/uuid"
)

// redis紀錄某人連線資訊
func (l *Logic) Connect(c context.Context, server string, token []byte) (uid, key, name, roomID string, hb int64, err error) {
	var params struct {
		// 認證中心token
		Token string `json:"token"`

		// client要進入的room
		RoomID string `json:"room_id"`
	}
	if err = json.Unmarshal(token, &params); err != nil {
		log.Errorf("json.Unmarshal(%s) error(%v)", token, err)
		return
	}
	roomID = params.RoomID

	// 告知comet連線多久沒心跳就直接close
	hb = l.c.Heartbeat

	key = uuid.New().String()

	uid, name = renew(params.Token)

	// 儲存user資料至redis
	if err = l.dao.AddMapping(c, uid, key, roomID, name, server); err != nil {
		log.Errorf("l.dao.AddMapping(%s,%s,%s,%s) error(%v)", uid, key, name, server, err)
	}
	log.Infof("conn connected key:%s server:%s uid:%s token:%s", key, server, uid, token)
	return
}

// redis清除某人連線資訊
func (l *Logic) Disconnect(c context.Context, uid, key, server string) (has bool, err error) {
	if has, err = l.dao.DelMapping(c, uid, key, server); err != nil {
		log.Errorf("l.dao.DelMapping(%s,%s,%s) error(%v)", uid, key, server, err)
		return
	}
	log.Infof("conn disconnected server:%s uid:%s key:%s", server, uid, key)
	return
}

// user key更換房間
func (l *Logic) ChangeRoom(c context.Context, uid, key, roomId string) (err error) {
	if err = l.dao.ChangeRoom(c, uid, key, roomId); err != nil {
		log.Errorf("l.dao.DelMapping(%s,%s)", uid, key)
		return
	}
	log.Infof("conn ChangeRoom  uid:%s key:%s roomId:%s", uid, key, roomId)
	return
}

// 更新某人redis資訊的過期時間
func (l *Logic) Heartbeat(c context.Context, uid, key, roomId, name, server string) (err error) {
	has, err := l.dao.ExpireMapping(c, uid)
	if err != nil {
		log.Errorf("l.dao.ExpireMapping(%s,%s,%s) error(%v)", uid, key, server, err)
		return
	}
	// 沒更新成功就直接做覆蓋
	if !has {
		if err = l.dao.AddMapping(c, uid, key, roomId, name, server); err != nil {
			log.Errorf("l.dao.AddMapping(%s,%s,%s) error(%v)", uid, key, server, err)
			return
		}
	}
	log.Infof("conn heartbeat key:%s server:%s uid:%d", key, server, uid)
	return
}

// restart redis內存的每個房間總人數
func (l *Logic) RenewOnline(c context.Context, server string, roomCount map[string]int32) (map[string]int32, error) {
	online := &dao.Online{
		Server:    server,
		RoomCount: roomCount,
		Updated:   time.Now().Unix(),
	}
	if err := l.dao.AddServerOnline(context.Background(), server, online); err != nil {
		return nil, err
	}
	return l.roomCount, nil
}
