package logic

import (
	"context"
	"encoding/json"
	"time"

	log "github.com/golang/glog"
	"github.com/google/uuid"
	"gitlab.com/jetfueltw/cpw/alakazam/internal/logic/model"
)

// redis紀錄某人連線資訊
func (l *Logic) Connect(c context.Context, server, cookie string, token []byte) (key, roomID string, hb int64, err error) {
	var params struct {
		// client key
		Key string `json:"key"`

		// client要進入的room
		RoomID string `json:"room_id"`

		// 裝置種類
		Platform string `json:"platform"`
	}
	if err = json.Unmarshal(token, &params); err != nil {
		log.Errorf("json.Unmarshal(%s) error(%v)", token, err)
		return
	}
	roomID = params.RoomID

	// 告知comet連線多久沒心跳就直接close
	hb = l.c.Heartbeat

	if key = params.Key; key == "" {
		key = uuid.New().String()
	}

	// 儲存user資料至redis
	if err = l.dao.AddMapping(c, key, server); err != nil {
		log.Errorf("l.dao.AddMapping(%s,%s) error(%v)", key, server, err)
	}
	log.Infof("conn connected key:%s server:%s token:%s", key, server, token)
	return
}

// redis清除某人連線資訊
func (l *Logic) Disconnect(c context.Context, key, server string) (has bool, err error) {
	if has, err = l.dao.DelMapping(c, key, server); err != nil {
		log.Errorf("l.dao.DelMapping(%s) error(%v)", key, server)
		return
	}
	log.Infof("conn disconnected key:%s server:%s", key, server)
	return
}

// 更新某人redis資訊的過期時間
func (l *Logic) Heartbeat(c context.Context, key, server string) (err error) {
	has, err := l.dao.ExpireMapping(c, key)
	if err != nil {
		log.Errorf("l.dao.ExpireMapping(%s,%s) error(%v)", key, server, err)
		return
	}
	// 沒更新成功就直接做覆蓋
	if !has {
		if err = l.dao.AddMapping(c, key, server); err != nil {
			log.Errorf("l.dao.AddMapping(%s,%s) error(%v)", key, server, err)
			return
		}
	}
	log.Infof("conn heartbeat key:%s server:%s", key, server)
	return
}

// restart redis內存的每個房間總人數
func (l *Logic) RenewOnline(c context.Context, server string, roomCount map[string]int32) (map[string]int32, error) {
	online := &model.Online{
		Server:    server,
		RoomCount: roomCount,
		Updated:   time.Now().Unix(),
	}
	if err := l.dao.AddServerOnline(context.Background(), server, online); err != nil {
		return nil, err
	}
	return l.roomCount, nil
}