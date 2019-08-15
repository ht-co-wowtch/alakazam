package member

import (
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/cache"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
)

type Member struct {
	db *models.Store

	cache *cache.Cache
}

func New(db *models.Store, cache *cache.Cache) *Member {
	return &Member{
		db:    db,
		cache: cache,
	}
}

// 一個聊天室會員的基本資料
type User struct {
	// user uid
	Uid string `json:"uid" binding:"required,len=32"`

	// user connection key
	Key string `json:"key" binding:"required,len=36"`

	// 房間id
	RoomId string `json:"-"`

	// 名稱
	Name string `json:"-"`

	// 權限狀態
	Status int `json:"-"`

	// 房間權限狀態
	RoomStatus int `json:"-"`
}

// 會員在線認證
func (l *Member) Auth(u *User) error {
	var err error
	u.RoomId, u.Name, u.Status, err = l.cache.GetUser(u.Uid, u.Key)
	if err != nil {
		return err
	}
	if u.Name == "" {
		return errors.ErrLogin
	}
	return nil
}
