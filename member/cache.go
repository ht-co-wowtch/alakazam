package member

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"time"
)

const (
	uidKey    = "uid_%s"
	uidWsKey  = "uid_ws_%s"
	bannedKey = "b_%s"

	OK = "OK"
)

type Cache struct {
	c      *redis.Client
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

var (
	errSetUidKey   = errors.New("set uid key")
	errSetUidWsKey = errors.New("set uid ws key")
	errExpire      = errors.New("set expire")
)

func (c *Cache) login(member *models.Member, key, server string) error {
	b, err := json.Marshal(member)
	if err != nil {
		return err
	}

	tx := c.c.Pipeline()

	c1 := tx.Set(keyUid(member.Uid), b, c.expire)

	uidWsKey := keyUidWs(member.Uid)
	c2 := tx.HSet(uidWsKey, key, server)
	c3 := tx.Expire(uidWsKey, c.expire)

	_, err = tx.Exec()
	if err != nil {
		return err
	}
	if c1.Val() != OK {
		return errSetUidKey
	}
	if !c2.Val() {
		return errSetUidWsKey
	}
	if !c3.Val() {
		return errExpire
	}
	return err
}

func (c *Cache) set(member *models.Member) (bool, error) {
	b, err := json.Marshal(member)
	if err != nil {
		return false, err
	}
	ok, err := c.c.Set(keyUid(member.Uid), b, c.expire).Result()
	return ok == OK, err
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

func (c *Cache) logout(uid, key string) (bool, error) {
	aff, err := c.c.HDel(keyUidWs(uid), key).Result()
	return aff == 1, err
}

func (c *Cache) delete(uid string) (bool, error) {
	tx := c.c.Pipeline()
	d1 := tx.Del(keyUid(uid))
	d2 := tx.Del(keyUidWs(uid))
	_, err := tx.Exec()
	if err != nil {
		return false, err
	}
	if d1.Val() != 1 {
		return false, nil
	}
	if d2.Val() != 1 {
		return false, nil
	}
	return true, nil
}

func (c *Cache) refreshExpire(uid string) error {
	tx := c.c.Pipeline()
	tx.Expire(keyUid(uid), c.expire)
	tx.Expire(keyUidWs(uid), c.expire)
	_, err := tx.Exec()
	return err
}

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

func (c *Cache) setBanned(uid string, expired time.Duration) (bool, error) {
	ok, err := c.c.Set(keyBanned(uid), time.Now().Add(expired).Unix(), expired).Result()
	if err != nil {
		return false, err
	}
	return ok == OK, nil
}

func (c *Cache) isBanned(uid string) (bool, error) {
	aff, err := c.c.Exists(keyBanned(uid)).Result()
	return aff == 1, err
}

func (c *Cache) delBanned(uid string) (bool, error) {
	aff, err := c.c.Del(keyBanned(uid)).Result()
	return aff == 1, err
}
