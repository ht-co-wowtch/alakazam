package member

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"gitlab.com/jetfueltw/cpw/micro/errdefs"
	"strconv"
	"time"
)

const (
	// user id的前綴詞，用於存儲在redis當key
	uidKey = "uid_%s"

	// user 禁言key的前綴詞
	bannedKey = "b_%s"

	// hash table json key
	hJsonKey = "data"

	// user hash table server key
	hServerKey = "server"

	// user hash table status key
	hashStatusKey = "status"
)

type Cache struct {
	c *redis.Client

	expire time.Duration
}

func newCache(client *redis.Client) *Cache {
	return &Cache{
		c:      client,
		expire: time.Minute * 30,
	}
}

func keyUid(uid string) string {
	return fmt.Sprintf(uidKey, uid)
}

func keyBanned(uid string) string {
	return fmt.Sprintf(bannedKey, uid)
}

// 儲存user資訊
func (c *Cache) login(member *models.Member, key, roomId, server string) error {
	b, err := json.Marshal(member)
	if err != nil {
		return err
	}

	keyI := keyUid(member.Uid)
	tx := c.c.Pipeline()
	f := map[string]interface{}{
		key:        roomId,
		hServerKey: server,
		hJsonKey:   b,
	}
	tx.HMSet(keyI, f)
	tx.Expire(keyI, c.expire)
	_, err = tx.Exec()
	return err
}

func (c *Cache) set(member *models.Member) error {
	b, err := json.Marshal(member)
	if err != nil {
		return err
	}
	return c.c.HSet(keyUid(member.Uid), hJsonKey, b).Err()
}

var errUserNil = errors.New("get user cache data has nil")

func (c *Cache) get(uid string) (*models.Member, error) {
	b, err := c.c.HGet(keyUid(uid), hJsonKey).Bytes()
	if err != nil {
		return nil, err
	}
	var m models.Member
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

type HMember struct {
	Mid       int
	Room      int
	Name      string
	Type      int
	IsMessage bool
}

func (c *Cache) getSession(uid string, key string) (HMember, error) {
	res, err := c.c.HMGet(keyUid(uid), key, hJsonKey).Result()
	if err != nil {
		return HMember{}, err
	}
	for _, v := range res {
		if v == nil {
			return HMember{}, errdefs.InvalidParameter(errUserNil, 1)
		}
	}

	var m models.Member
	if err = json.Unmarshal([]byte(res[1].(string)), &m); err != nil {
		return HMember{}, err
	}

	rid, err := strconv.Atoi(res[0].(string))
	if err != nil {
		return HMember{}, err
	}
	return HMember{
		Mid:       m.Id,
		Room:      rid,
		Name:      m.Name,
		Type:      m.Type,
		IsMessage: m.IsMessage,
	}, nil
}

// 移除user資訊
// DEL : uid_{user id}
func (c *Cache) delete(uid, key string) (bool, error) {
	aff, err := c.c.HDel(keyUid(uid), key).Result()
	return aff >= 1, err
}

// restart user資料的過期時間
// EXPIRE : uid_{user id}  (HSET)
func (c *Cache) refreshExpire(uid string) (bool, error) {
	return c.c.Expire(keyUid(uid), c.expire).Result()
}

// 取會員名稱
func (c *Cache) getName(uid []string) ([]string, error) {
	tx := c.c.Pipeline()
	cmd := make([]*redis.StringCmd, len(uid))
	for i, id := range uid {
		cmd[i] = tx.HGet(keyUid(id), hJsonKey)
	}
	_, err := tx.Exec()
	if err != nil {
		return nil, err
	}
	name := make([]string, len(uid))
	for i, v := range cmd {
		var m models.Member
		b, _ := v.Bytes()
		if err = json.Unmarshal(b, &m); err != nil {
			return nil, err
		}
		name[i] = m.Name
	}
	return name, nil
}

// 設定禁言
func (c *Cache) setBanned(uid string, expired time.Duration) error {
	key := keyBanned(uid)
	ok, err := c.c.SetNX(key, time.Now().Add(expired).Unix(), expired).Result()
	if err != nil {
		return err
	}
	if ok {
		m, err := c.get(uid)
		if err != nil {
			return err
		}
		m.IsMessage = false
		return c.set(m)
	}
	return nil
}

// 取得禁言時效
func (c *Cache) isBanned(uid string) (bool, error) {
	i, err := c.c.Exists(keyBanned(uid)).Result()
	return i >= 1, err
}

// 解除禁言
func (c *Cache) delBanned(uid string) error {
	_, err := c.c.Del(keyBanned(uid)).Result()
	if err != nil {
		return err
	}
	m, err := c.get(uid)
	if err != nil {
		return err
	}
	m.IsMessage = true
	return c.set(m)
}

// 更換房間
func (c *Cache) changeRoom(uid, key, roomId string) error {
	tx := c.c.Pipeline()
	tx.HSet(keyUid(uid), key, roomId)
	tx.Expire(keyUid(uid), c.expire)
	_, err := tx.Exec()
	return err
}
