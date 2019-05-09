package logic

import (
	"context"
	"encoding/json"
	"gitlab.com/jetfueltw/cpw/alakazam/internal/logic/dao"
	"time"

	log "github.com/golang/glog"
	"github.com/google/uuid"
)

// redis紀錄某人連線資訊
func (l *Logic) Connect(c context.Context, server, cookie string, token []byte) (mid int64, key, name, roomID string, hb int64, err error) {
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

	mid, name = renew(params.Token)

	// 儲存user資料至redis
	if err = l.dao.AddMapping(c, mid, key, name, server); err != nil {
		log.Errorf("l.dao.AddMapping(%d,%s,%s,%s) error(%v)", mid, key, name, server, err)
	}
	log.Infof("conn connected key:%s server:%s mid:%d token:%s", key, server, mid, token)
	return
}

// redis清除某人連線資訊
func (l *Logic) Disconnect(c context.Context, mid int64, key, server string) (has bool, err error) {
	if has, err = l.dao.DelMapping(c, mid, key, server); err != nil {
		log.Errorf("l.dao.DelMapping(%d,%s,%s) error(%v)", mid, key, server, err)
		return
	}
	log.Infof("conn disconnected key:%s server:%s mid:%d", key, server, mid)
	return
}

// 更新某人redis資訊的過期時間
func (l *Logic) Heartbeat(c context.Context, mid int64, key, name, server string) (err error) {
	has, err := l.dao.ExpireMapping(c, mid)
	if err != nil {
		log.Errorf("l.dao.ExpireMapping(%d,%s,%s) error(%v)", mid, key, server, err)
		return
	}
	// 沒更新成功就直接做覆蓋
	if !has {
		if err = l.dao.AddMapping(c, mid, key, name, server); err != nil {
			log.Errorf("l.dao.AddMapping(%d,%s,%s) error(%v)", mid, key, server, err)
			return
		}
	}
	log.Infof("conn heartbeat key:%s server:%s mid:%d", key, server, mid)
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
