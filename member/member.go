package member

import (
	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"gitlab.com/jetfueltw/cpw/alakazam/client"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
)

type Member struct {
	cli *client.Client
	db  *models.Store
	c   *Cache
}

func New(db *models.Store, cache *redis.Client, cli *client.Client) *Member {
	return &Member{
		db:  db,
		cli: cli,
		c:   newCache(cache),
	}
}

// 一個聊天室會員的基本資料
type User struct {
	Uid        string  `json:"-"`
	RoomId     string  `json:"room_id"`
	Key        string  `json:"key" binding:"required,len=36"`
	RoomStatus int     `json:"-"`
	H          HMember `json:"-"`
}

func (m *Member) Login(token, roomId, server string) (*models.Member, string, error) {
	user, err := m.cli.Auth(token)
	if err != nil {
		return nil, "", err
	}

	u, ok, err := m.db.Find(user.Uid)
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
		if aff, err := m.db.CreateUser(u); err != nil || aff <= 0 {
			return nil, "", err
		}
	} else if u.IsBlockade {
		return u, "", nil
	}

	if u.Name != user.Name || u.Avatar != user.Avatar {
		u.Name = user.Name
		u.Avatar = user.Avatar
		if aff, err := m.db.UpdateUser(u); err != nil || aff <= 0 {
			log.Error("UpdateUser", zap.String("uid", user.Uid), zap.Int64("affected", aff), zap.Error(err))
		}
	}

	key := uuid.New().String()

	// 儲存user資料至redis
	if err := m.c.login(u, key, server); err != nil {
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

func (m *Member) Logout(uid, key string) (bool, error) {
	return m.c.logout(uid, key)
}

func (m *Member) Kick(uid string) ([]string, error) {
	keys, err := m.c.getKey(uid)
	if err != nil {
		return nil, err
	}
	err = m.c.delete(uid)
	return keys, err
}

func (m *Member) GetKeys(uid string) ([]string, error) {
	return m.c.getKey(uid)
}

func (m *Member) Get(uid string) (*models.Member, error) {
	return m.c.get(uid)
}

func (m *Member) GetUserName(uid []string) ([]string, error) {
	name, err := m.c.getName(uid)
	if err == redis.Nil {
		return nil, errors.ErrNoRows
	}
	return name, nil
}

func (m *Member) GetMembers(id []int) ([]models.Member, error) {
	return m.db.GetMembers(id)
}

func (m *Member) Heartbeat(uid string) error {
	err := m.c.refreshExpire(uid)
	if err != nil {
		return err
	}
	return nil
}
