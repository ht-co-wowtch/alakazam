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
	uidKey    = "uid_h_%s"
	uidWsKey  = "uid_h_ws_%s"
	bannedKey = "b_%s"

	uidJsonKey = "json"
	uidNameKey = "name"

	OK = "OK"
)

type Cache interface {
	login(member *models.Member, key, server string) error
	set(member *models.Member) (bool, error)
	get(uid string) (*models.Member, error)
	getKey(uid string) ([]string, error)
	logout(uid, key string) (bool, error)
	delete(uid string) (bool, error)
	refreshExpire(uid string) error
	setName(name map[string]string) error
	getName(uid []string) (map[string]string, error)
	setBanned(uid string, expired time.Duration) (bool, error)
	isBanned(uid string) (bool, error)
	delBanned(uid string) (bool, error)
}

type cache struct {
	c      *redis.Client
	expire time.Duration
}

func newCache(client *redis.Client) Cache {
	return &cache{
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

func (c *cache) login(member *models.Member, key, server string) error {
	b, err := json.Marshal(member)
	if err != nil {
		return err
	}

	tx := c.c.Pipeline()

	uidKey := keyUid(member.Uid)
	c1 := tx.HMSet(uidKey, map[string]interface{}{
		uidJsonKey: b,
		uidNameKey: member.Name,
	})

	uidWsKey := keyUidWs(member.Uid)
	c2 := tx.HSet(uidWsKey, key, server)
	c3 := tx.Expire(uidKey, c.expire)
	c4 := tx.Expire(uidWsKey, c.expire)

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
	if !c3.Val() || !c4.Val() {
		return errExpire
	}
	return err
}

func (c *cache) set(member *models.Member) (bool, error) {
	b, err := json.Marshal(member)
	if err != nil {
		return false, err
	}
	return c.c.HSet(keyUid(member.Uid), uidJsonKey, b).Result()
}

func (c *cache) get(uid string) (*models.Member, error) {
	b, err := c.c.HGet(keyUid(uid), uidJsonKey).Bytes()
	if err != nil {
		return nil, err
	}
	var m models.Member
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (c *cache) getKey(uid string) ([]string, error) {
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

func (c *cache) logout(uid, key string) (bool, error) {
	aff, err := c.c.HDel(keyUidWs(uid), key).Result()
	return aff == 1, err
}

func (c *cache) delete(uid string) (bool, error) {
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

func (c *cache) refreshExpire(uid string) error {
	tx := c.c.Pipeline()
	tx.Expire(keyUid(uid), c.expire)
	tx.Expire(keyUidWs(uid), c.expire)
	_, err := tx.Exec()
	return err
}

func (c *cache) setName(name map[string]string) error {
	tx := c.c.Pipeline()
	for id, na := range name {
		tx.HSet(keyUid(id), uidNameKey, na)
	}
	_, err := tx.Exec()
	return err
}

func (c *cache) getName(uid []string) (map[string]string, error) {
	tx := c.c.Pipeline()
	cmd := make(map[string]*redis.StringCmd, len(uid))
	for _, id := range uid {
		cmd[id] = tx.HGet(keyUid(id), uidNameKey)
	}
	_, err := tx.Exec()
	if err != nil {
		return nil, err
	}
	name := make(map[string]string, len(uid))
	for id, na := range cmd {
		name[id] = na.Val()
	}
	return name, nil
}

func (c *cache) setBanned(uid string, expired time.Duration) (bool, error) {
	ok, err := c.c.Set(keyBanned(uid), time.Now().Add(expired).Unix(), expired).Result()
	if err != nil {
		return false, err
	}
	return ok == OK, nil
}

func (c *cache) isBanned(uid string) (bool, error) {
	aff, err := c.c.Exists(keyBanned(uid)).Result()
	return aff == 1, err
}

func (c *cache) delBanned(uid string) (bool, error) {
	aff, err := c.c.Del(keyBanned(uid)).Result()
	return aff == 1, err
}
