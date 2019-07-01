package logic

import (
	log "github.com/golang/glog"
	"github.com/google/uuid"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
)

// 一個聊天室會員的基本資料
type User struct {
	// user uid
	Uid string `json:"uid" binding:"required"`

	// user connection key
	Key string `json:"key" binding:"required"`

	// 房間id
	RoomId string `json:"-"`

	// 名稱
	name string `json:"-"`

	// 權限狀態
	status int `json:"-"`

	// 房間權限狀態
	roomStatus int `json:"-"`
}

// 會員在線認證
func (l *Logic) auth(u *User) (err error) {
	u.RoomId, u.name, u.status, err = l.cache.GetUser(u.Uid, u.Key)

	if err != nil {
		return errors.FailureError
	}

	if u.name == "" {
		return errors.LoginError
	}

	if u.RoomId == "" {
		return errors.RoomError
	}

	return nil
}

// 房間權限認證
func (l *Logic) authRoom(u *User) error {
	u.roomStatus = l.GetRoomPermission(u.RoomId)

	if models.IsBanned(u.roomStatus) {
		return errors.RoomBannedError
	}

	if l.isUserBanned(u.Uid, u.status) {
		return errors.BannedError
	}

	return nil
}

// 登入到聊天室
func (l *Logic) login(token, roomId, server string) (*models.Member, string, error) {
	user, err := l.client.Auth(token)
	if err != nil {
		log.Errorf("Logic client GetUser token:%s error(%v)", token, err)
		return nil, "", errors.UserError
	}

	// TODO 處理error
	u, ok, _ := l.db.Find(user.Uid)
	if !ok {
		u = &models.Member{
			Uid:    user.Uid,
			Name:   user.Name,
			Avatar: user.Avatar,
			Type:   user.Type,
		}
		if aff, err := l.db.CreateUser(u); err != nil || aff <= 0 {
			log.Errorf("CreateUser(uid:%s) affected %d error(%v)", user.Uid, aff, err)
			return nil, "", errors.ConnectError
		}
	} else if u.IsBlockade {
		return u, "", errors.BlockadeError
	}

	if u.Name != user.Name || u.Avatar != user.Avatar {
		u.Name = user.Name
		u.Avatar = user.Avatar
		if aff, err := l.db.UpdateUser(u); err != nil || aff <= 0 {
			log.Errorf("UpdateUser(uid:%s) affected %d error(%v)", user.Uid, aff, err)
		}
	}

	key := uuid.New().String()

	// 儲存user資料至redis
	if err := l.cache.SetUser(u, key, roomId, server); err != nil {
		log.Errorf("l.dao.SetUser(%s,%s,%s,%s) error(%v)", u.Uid, key, u.Name, server, err)
	} else {
		log.Infof("conn connected key:%s server:%s uid:%s token:%s", key, server, u.Uid, token)
	}
	return u, key, nil
}

func (l *Logic) GetUserName(uid []string) ([]string, error) {
	return l.cache.GetUserName(uid)
}
