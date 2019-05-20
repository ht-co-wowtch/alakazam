package logic

import (
	"database/sql"
	"encoding/json"
	"github.com/google/uuid"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/business"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/cache"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/remote"
	"time"

	log "github.com/golang/glog"
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
		// 認證中心token
		Token string `json:"token"`

		// client要進入的room
		RoomID string `json:"room_id"`
	}
	if err := json.Unmarshal(token, &params); err != nil {
		log.Errorf("json.Unmarshal(%s) error(%v)", token, err)
		return nil, errors.ConnectError
	}

	r := new(ConnectReply)
	r.Uid, r.Name = remote.Renew(params.Token)
	permission, isBlockade, err := l.db.FindUserPermission(r.Uid)

	if err != nil {
		if err != sql.ErrNoRows {
			log.Errorf("FindUserPermission(uid:%s) error(%v)", r.Uid, err)
			return r, errors.ConnectError
		}
		if aff, err := l.db.CreateUser(r.Uid, business.PlayDefaultPermission); err != nil || aff <= 0 {
			log.Errorf("CreateUser(uid:%s) affected %d error(%v)", r.Uid, aff, err)
			return r, errors.ConnectError
		}
		permission = business.PlayDefaultPermission
	} else if isBlockade {
		r.Permission = business.Blockade
		log.Infof("conn blockade uid:%s token:%s", r.Uid, token)
		return r, nil
	}

	r.Permission = permission
	r.RoomId = params.RoomID

	// 告知comet連線多久沒心跳就直接close
	r.Hb = l.c.Heartbeat

	r.Key = uuid.New().String()

	// 儲存user資料至redis
	if err := l.cache.AddMapping(r.Uid, r.Key, r.RoomId, r.Name, server, r.Permission); err != nil {
		log.Errorf("l.dao.AddMapping(%s,%s,%s,%s) error(%v)", r.Uid, r.Key, r.Name, server, err)
	}
	log.Infof("conn connected key:%s server:%s uid:%s token:%s", r.Key, server, r.Uid, token)
	return r, nil
}

// redis清除某人連線資訊
func (l *Logic) Disconnect(uid, key, server string) (has bool, err error) {
	if has, err = l.cache.DelMapping(uid, key, server); err != nil {
		log.Errorf("l.dao.DelMapping(%s,%s,%s) error(%v)", uid, key, server, err)
		return
	}
	log.Infof("conn disconnected server:%s uid:%s key:%s", server, uid, key)
	return
}

// user key更換房間
func (l *Logic) ChangeRoom(uid, key, roomId string) (err error) {
	if err = l.cache.ChangeRoom(uid, key, roomId); err != nil {
		log.Errorf("l.dao.DelMapping(%s,%s)", uid, key)
		return
	}
	log.Infof("conn ChangeRoom  uid:%s key:%s roomId:%s", uid, key, roomId)
	return
}

// 更新某人redis資訊的過期時間
func (l *Logic) Heartbeat(uid, key, roomId, name, server string) (err error) {
	has, err := l.cache.ExpireMapping(uid)
	if err != nil {
		log.Errorf("l.dao.ExpireMapping(%s,%s,%s) error(%v)", uid, key, server, err)
		return
	}
	// 沒更新成功就直接做覆蓋
	if !has {
		// TODO 要重抓user 權限值帶到status欄位
		if err = l.cache.AddMapping(uid, key, roomId, name, server, 0); err != nil {
			log.Errorf("l.dao.AddMapping(%s,%s,%s) error(%v)", uid, key, server, err)
			return
		}
	}
	log.Infof("conn heartbeat key:%s server:%s uid:%s", key, server, uid)
	return
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
