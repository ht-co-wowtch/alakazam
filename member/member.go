package member

import (
	"database/sql"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"gitlab.com/jetfueltw/cpw/alakazam/client"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
)

type Chat interface {
	GetSession(uid string) (*models.Member, error)
	Login(rid int, token, server string) (*models.Member, string, error)
	Logout(uid, key string) (bool, error)
	Heartbeat(uid string) error
}

type Member struct {
	cli *client.Client
	db  models.Chat
	c   Cache
}

func New(db models.Chat, cache *redis.Client, cli *client.Client) *Member {
	return &Member{
		db:  db,
		cli: cli,
		c:   newCache(cache),
	}
}

var (
	errInsertMember = errors.New("insert member")
)

func (m *Member) Login(rid int, token, server string) (*models.Member, string, error) {
	user, err := m.cli.Auth(token)
	if err != nil {
		return nil, "", err
	}

	u, err := m.db.Find(user.Uid)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, "", err
		}

		u = &models.Member{
			Uid:    user.Uid,
			Name:   user.Name,
			Avatar: user.Avatar,
			Type:   user.Type,
		}
		ok, err := m.db.CreateUser(u)
		if err != nil {
			return nil, "", err
		}
		if !ok {
			return nil, "", errInsertMember
		}
	}
	if u.IsBlockade {
		return u, "", nil
	}

	if u.Name != user.Name || u.Avatar != user.Avatar {
		u.Name = user.Name
		u.Avatar = user.Avatar
		if ok, err := m.db.UpdateUser(u); err != nil || !ok {
			log.Error("update user", zap.String("uid", user.Uid), zap.Bool("action", ok), zap.Error(err))
		}
	}

	key := uuid.New().String()

	if err = m.c.login(u, key, server); err != nil {
		return nil, "", err
	} else {
		log.Info(
			"conn connected",
			zap.String("key", key),
			zap.String("uid", u.Uid),
			zap.Int("room_id", rid),
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
	ok, err := m.c.delete(uid)
	if !ok {
		// TODO error
		return nil, err
	}
	return keys, err
}

func (m *Member) GetKeys(uid string) ([]string, error) {
	return m.c.getKey(uid)
}

func (m *Member) GetMessageSession(uid string) (*models.Member, error) {
	member, err := m.c.get(uid)
	if err != nil {
		if err == redis.Nil {
			return nil, errors.ErrLogin
		}
		return nil, err
	}
	if member.Type == models.Guest {
		return nil, errors.ErrLogin
	}
	if !member.IsMessage {
		return nil, errors.ErrMemberNoMessage
	}

	ok, err := m.c.isBanned(uid)
	if err != nil {
		return nil, err
	}
	if ok {
		return nil, errors.ErrMemberBanned
	}
	return member, nil
}

func (m *Member) GetSession(uid string) (*models.Member, error) {
	member, err := m.c.get(uid)
	if err != nil {
		if err == redis.Nil {
			return nil, errors.ErrLogin
		}
		return nil, err
	}
	if member.Type == models.Guest {
		return nil, errors.ErrLogin
	}
	return member, nil
}

func (m *Member) GetUserName(uid []string) (map[string]string, error) {
	name, err := m.c.getName(uid)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	selectUid := make([]string, 0)
	for _, id := range uid {
		if _, ok := name[id]; !ok {
			selectUid = append(selectUid, id)
		}
	}

	member, err := m.db.GetMembersByUid(selectUid)
	if err != nil {
		return nil, err
	}
	if name == nil {
		name = make(map[string]string, len(member))
	}

	cacheName := make(map[string]string, len(member))
	for _, v := range member {
		cacheName[v.Uid] = v.Name
		name[v.Uid] = v.Name
	}
	if err := m.c.setName(cacheName); err != nil {
		log.Error("set name cache for GetUserName", zap.Error(err), zap.Any("name", name))
	}
	return name, nil
}

func (m *Member) GetMembers(id []int) ([]models.Member, error) {
	return m.db.GetMembers(id)
}

func (m *Member) Heartbeat(uid string) error {
	return m.c.refreshExpire(uid)
}
