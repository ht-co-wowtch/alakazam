package member

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"time"
)

const (
	// user id的前綴詞，用於存儲在redis當key
	uidKey = "uid_%s"

	uidWsKey = "uid_ws_%s"

	// user 禁言key的前綴詞
	bannedKey = "b_%s"
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

func keyUidWs(uid string) string {
	return fmt.Sprintf(uidWsKey, uid)
}

func keyBanned(uid string) string {
	return fmt.Sprintf(bannedKey, uid)
}

// 儲存user資訊
func (c *Cache) login(member *models.Member, key, server string) error {
	b, err := json.Marshal(member)
	if err != nil {
		return err
	}

	tx := c.c.Pipeline()

	tx.Set(keyUid(member.Uid), b, c.expire)

	uidWsKey := keyUidWs(member.Uid)
	tx.HSet(uidWsKey, key, server)
	tx.Expire(uidWsKey, c.expire)

	_, err = tx.Exec()
	return err
}

func (c *Cache) set(member *models.Member) error {
	b, err := json.Marshal(member)
	if err != nil {
		return err
	}
	return c.c.Set(keyUid(member.Uid), b, c.expire).Err()
}

func (c *Cache) get(uid string) (*models.Member, error) {
	b, err := c.c.Get(keyUid(uid)).Bytes()
	if err != nil {
		return nil, err
	}
	var m models.Member
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (c *Cache) getKey(uid string) ([]string, error) {
	maps, err := c.c.HGetAll(keyUidWs(uid)).Result()
	if err != nil {
		return nil, err
	}
	var keys []string
	for key, _ := range maps {
		keys = append(keys, key)
	}
	return keys, nil
}

type HMember struct {
	Mid       int
	Room      int
	Name      string
	Type      int
	IsMessage bool
}

// 移除user資訊
// DEL : uid_{user id}
func (c *Cache) logout(uid, key string) (bool, error) {
	aff, err := c.c.HDel(keyUidWs(uid), key).Result()
	return aff >= 1, err
}

func (c *Cache) delete(uid string) error {
	tx := c.c.Pipeline()
	tx.Del(keyUid(uid))
	tx.Del(keyUidWs(uid))
	_, err := tx.Exec()
	return err
}

// restart user資料的過期時間
// EXPIRE : uid_{user id}  (HSET)
func (c *Cache) refreshExpire(uid string) error {
	tx := c.c.Pipeline()
	tx.Expire(keyUid(uid), c.expire)
	tx.Expire(keyUidWs(uid), c.expire)
	_, err := tx.Exec()
	return err
}

// 取會員名稱
func (c *Cache) getName(uid []string) ([]string, error) {
	tx := c.c.Pipeline()
	cmd := make([]*redis.StringCmd, len(uid))
	for i, id := range uid {
		cmd[i] = tx.Get(keyUid(id))
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
	_, err := c.c.Set(key, time.Now().Add(expired).Unix(), expired).Result()
	if err != nil {
		return err
	}
	return nil
}

// 是否禁言中
func (c *Cache) isBanned(uid string) (bool, error) {
	i, err := c.c.Exists(keyBanned(uid)).Result()
	return i >= 1, err
}

// 解除禁言
func (c *Cache) delBanned(uid string) error {
	_, err := c.c.Del(keyBanned(uid)).Result()
	return err
}
