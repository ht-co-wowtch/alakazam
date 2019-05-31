package logic

import (
	"database/sql"
	log "github.com/golang/glog"
	"github.com/google/uuid"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/permission"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/store"
)

// 一個聊天室會員的基本資料
type User struct {
	// user uid
	Uid string `json:"uid" binding:"required"`

	// user connection key
	Key string `json:"key" binding:"required"`

	// 房間id
	roomId string

	// 名稱
	name string

	// 權限狀態
	status int

	// 房間權限狀態
	roomStatus int
}

// 會員在線認證
func (l *Logic) auth(u *User) (err error) {
	u.roomId, u.name, u.status, err = l.cache.GetUser(u.Uid, u.Key)

	if err != nil {
		return errors.FailureError
	}

	if u.name == "" {
		return errors.LoginError
	}

	if u.roomId == "" {
		return errors.RoomError
	}

	return nil
}

// 房間權限認證
func (l *Logic) authRoom(u *User) error {
	u.roomStatus = l.GetRoomPermission(u.roomId)

	if permission.IsBanned(u.roomStatus) {
		return errors.RoomBannedError
	}

	if l.isUserBanned(u.Uid, u.status) {
		return errors.BannedError
	}

	return nil
}

// 登入到聊天室
func (l *Logic) login(token, roomId, server string) (*store.User, string, error) {
	user, err := l.client.Auth(token)
	if err != nil {
		log.Errorf("Logic client GetUser token:%s error(%v)", token, err)
		return nil, "", errors.UserError
	}

	u, err := l.db.Find(user.Uid)

	if err != nil {
		if err != sql.ErrNoRows {
			log.Errorf("FindUserPermission(uid:%s) error(%v)", user.Uid, err)
			return nil, "", errors.ConnectError
		}

		u = &store.User{
			Uid:        user.Uid,
			Name:       user.Nickname,
			Avatar:     user.Avatar,
			Permission: permission.PlayDefaultPermission,
		}

		if aff, err := l.db.CreateUser(u); err != nil || aff <= 0 {
			log.Errorf("CreateUser(uid:%s) affected %d error(%v)", user.Uid, aff, err)
			return nil, "", errors.ConnectError
		}
	} else if u.IsBlockade {
		return u, "", errors.BlockadeError
	}

	if u.Name != user.Nickname || u.Avatar != user.Avatar {
		u.Name = user.Nickname
		u.Avatar = user.Avatar
		if aff, err := l.db.UpdateUser(u); err != nil || aff <= 0 {
			log.Errorf("UpdateUser(uid:%s) affected %d error(%v)", user.Uid, aff, err)
		}
	}

	key := uuid.New().String()

	// 儲存user資料至redis
	if err := l.cache.SetUser(u.Uid, key, roomId, u.Name, server, u.Permission); err != nil {
		log.Errorf("l.dao.SetUser(%s,%s,%s,%s) error(%v)", u.Uid, key, u.Name, server, err)
	} else {
		log.Infof("conn connected key:%s server:%s uid:%s token:%s", key, server, u.Uid, token)
	}

	return u, key, nil
}
