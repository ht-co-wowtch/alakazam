package logic

import (
	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
)

// 一個聊天室會員的基本資料
type User struct {
	// user uid
	Uid string `json:"uid" binding:"required,len=32"`

	// user connection key
	Key string `json:"key" binding:"required,len=36"`

	// 房間id
	RoomId string `json:"-"`

	// 名稱
	name string `json:"-"`

	// 權限狀態
	Status int `json:"-"`

	// 房間權限狀態
	roomStatus int `json:"-"`
}

// 會員在線認證
func (l *Logic) auth(u *User) error {
	var err error
	u.RoomId, u.name, u.Status, err = l.cache.GetUser(u.Uid, u.Key)
	if err != nil {
		return err
	}
	if u.name == "" {
		return errors.ErrLogin
	}
	return nil
}

// 房間權限認證
func (l *Logic) authRoom(u *User) error {
	var err error
	u.roomStatus, err = l.cache.GetRoom(u.RoomId)
	if err != nil && err != redis.Nil {
		return err
	}
	if u.roomStatus == 0 {
		room, _, err := l.db.GetRoom(u.RoomId)
		if err != nil {
			return err
		}
		u.roomStatus = room.Status()
		if err := l.cache.SetRoom(room); err != nil {
			return err
		}
	}
	if models.IsBanned(u.roomStatus) {
		return errors.ErrRoomBanned
	}
	is, err := l.isUserBanned(u.Uid, u.Status)
	if err != nil {
		return err
	}
	if is {
		return errors.ErrBanned
	}
	return nil
}

// 登入到聊天室
func (l *Logic) login(token, roomId, server string) (*models.Member, string, error) {
	user, err := l.client.Auth(token)
	if err != nil {
		return nil, "", err
	}

	u, ok, err := l.db.Find(user.Uid)
	if err != nil {
		return nil, "", err
	}
	if !ok {
		u = &models.Member{
			Uid:    user.Uid,
			Name:   user.Name,
			Avatar: user.Avatar,
			Type:   user.Type,
		}
		if aff, err := l.db.CreateUser(u); err != nil || aff <= 0 {
			return nil, "", err
		}
	} else if u.IsBlockade {
		return u, "", nil
	}

	if u.Name != user.Name || u.Avatar != user.Avatar {
		u.Name = user.Name
		u.Avatar = user.Avatar
		if aff, err := l.db.UpdateUser(u); err != nil || aff <= 0 {
			log.Error("UpdateUser", zap.String("uid", user.Uid), zap.Int64("affected", aff), zap.Error(err))
		}
	}

	key := uuid.New().String()

	// 儲存user資料至redis
	if err := l.cache.SetUser(u, key, roomId, server); err != nil {
		return nil, "", err
	} else {
		log.Info(
			"conn connected",
			zap.String("key", key),
			zap.String("uid", u.Uid),
			zap.String("room_id", roomId),
			zap.String("server", server),
		)
	}
	return u, key, nil
}

func (l *Logic) GetUserName(uid []string) ([]string, error) {
	name, err := l.cache.GetUserName(uid)
	if err == redis.Nil {
		return nil, errors.ErrNoRows
	}
	return name, nil
}
