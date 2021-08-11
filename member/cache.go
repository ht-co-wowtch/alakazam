package member

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
)

const (
	prefix = "ala"

	uidWsKey  = prefix + ":uid_h_ws_%s"
	bannedKey = prefix + ":b_%s-%d"
	uidKey    = prefix + ":uid_h_%s"

	uidDataHKey = "data"

	uidNameHKey = "name"

	OK = "OK"
)

type Cache interface {
	login(member *models.Member, rid int, key string) error
	set(member *models.Member) error
	get(uid string) (*models.Member, error)
	getByRoom(uid string, rid int) (*models.Member, error)
	getKeys(uid string) ([]string, error)
	getRoomKeys(uid string, rid int) ([]string, error)
	getWs(uid string) (map[string]string, error)
	setWs(uid, key string, rid int) error
	logout(uid, key string) (bool, error)
	delete(uid string) (bool, error)
	clearRoom(uid string) error
	refreshExpire(uid string) error
	setName(name map[string]string) error
	getName(uid []string) (map[string]string, error)
	isBanned(uid string, rid int) (bool, error)
	setBanned(uid string, rid int, expired time.Duration) error
	delBanned(uid string, rid int) error

	delAllBanned(uid string) error
}

type cache struct {
	c      *redis.Client
	expire time.Duration
}

func keyUid(uid string) string {
	return fmt.Sprintf(uidKey, uid)
}

func keyUidWs(uid string) string {
	return fmt.Sprintf(uidWsKey, uid)
}

func NewCache(client *redis.Client) Cache {
	return &cache{
		c:      client,
		expire: time.Minute * 30,
	}
}

func keyBanned(uid string, rid int) string {
	return fmt.Sprintf(bannedKey, uid, rid)
}

var (
	errSetUidKey   = errors.New("set uid key")
	errSetUidWsKey = errors.New("set uid ws key")
	errExpire      = errors.New("set expire")
)

func (c *cache) login(member *models.Member, rid int, key string) error {
	b, err := json.Marshal(member)
	if err != nil {
		return err
	}

	isb, _ := json.Marshal([]bool{
		member.Permission.IsBanned,
		member.Permission.IsBlockade,
		member.Permission.IsManage,
	})

	tx := c.c.Pipeline()

	uidKey := keyUid(member.Uid)

	c1 := tx.HMSet(uidKey, map[string]interface{}{
		uidDataHKey:       b,
		uidNameHKey:       member.Name,
		strconv.Itoa(rid): isb,
	})

	uidWsKey := keyUidWs(member.Uid)
	c2 := tx.HSet(uidWsKey, key, rid)

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

func (c *cache) set(member *models.Member) error {
	b, err := json.Marshal(member)
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		uidDataHKey: b,
	}

	if member.Permission.RoomId != 0 {
		isb, _ := json.Marshal([]bool{
			member.Permission.IsBanned,
			member.Permission.IsBlockade,
			member.Permission.IsManage,
		})

		data[strconv.Itoa(int(member.Permission.RoomId))] = isb
	}

	key := keyUid(member.Uid)
	tx := c.c.Pipeline()
	tx.HMSet(key, data)
	tx.Expire(key, c.expire)

	_, err = tx.Exec()

	return err
}

func (c *cache) get(uid string) (*models.Member, error) {
	b, err := c.c.HGet(keyUid(uid), uidDataHKey).Bytes()
	if err != nil {
		return nil, err
	}

	m := new(models.Member)
	if err := json.Unmarshal(b, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *cache) getKeys(uid string) ([]string, error) {
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

func (c *cache) getRoomKeys(uid string, rid int) ([]string, error) {
	maps, err := c.c.HGetAll(keyUidWs(uid)).Result()
	if err != nil {
		return nil, err
	}

	sRid := strconv.Itoa(rid)
	var keys []string
	for key, id := range maps {
		if id == sRid {
			keys = append(keys, key)
		}
	}
	return keys, nil
}

func (c *cache) getByRoom(uid string, rid int) (*models.Member, error) {
	member, err := c.c.HMGet(keyUid(uid), uidDataHKey, strconv.Itoa(rid)).Result()
	if err != nil {
		return nil, err
	}

	if member[0] == nil {
		return nil, redis.Nil
	}

	m := new(models.Member)
	if err := json.Unmarshal([]byte(member[0].(string)), &m); err != nil {
		return nil, err
	}

	if member[1] == nil {
		return m, nil
	}

	var is []bool
	if err := json.Unmarshal([]byte(member[1].(string)), &is); err != nil {
		return nil, err
	}

	m.Permission.RoomId = int64(rid)
	m.Permission.IsBanned = is[0]
	m.Permission.IsBlockade = is[1]
	m.Permission.IsManage = is[2]

	return m, nil
}

func (c *cache) setWs(uid, key string, rid int) error {
	uidWsKey := keyUidWs(uid)
	c1 := c.c.HSet(uidWsKey, key, rid)
	c2 := c.c.Expire(uidWsKey, c.expire)

	if c1.Err() != nil {
		return errSetUidWsKey
	}
	if !c2.Val() {
		return errExpire
	}

	return nil
}

func (c *cache) getWs(uid string) (map[string]string, error) {
	return c.c.HGetAll(keyUidWs(uid)).Result()
}

// 刪除快取連線紀錄
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

// 刪除房間快取
func (c *cache) clearRoom(uid string) error {
	m, err := c.get(uid)
	if err != nil {
		return err
	}

	b, err := json.Marshal(m)
	if err != nil {
		return err
	}

	tx := c.c.Pipeline()
	uidKey := keyUid(m.Uid)
	tx.Del(uidKey)
	tx.HMSet(uidKey, map[string]interface{}{
		uidDataHKey: b,
		uidNameHKey: m.Name,
	})
	tx.Expire(uidKey, c.expire)

	_, err = tx.Exec()
	return err
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
		tx.HSet(keyUid(id), uidNameHKey, na)
	}
	_, err := tx.Exec()
	return err
}

func (c *cache) getName(uid []string) (map[string]string, error) {
	tx := c.c.Pipeline()
	cmd := make(map[string]*redis.StringCmd, len(uid))
	for _, id := range uid {
		cmd[id] = tx.HGet(keyUid(id), uidNameHKey)
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

func (c *cache) setBanned(uid string, rid int, expired time.Duration) error {
	_, err := c.c.Set(keyBanned(uid, rid), time.Now().Add(expired).Unix(), expired).Result()
	if err != nil {
		return err
	}
	return nil
}

func (c *cache) delBanned(uid string, rid int) error {
	_, err := c.c.Del(keyBanned(uid, rid)).Result()
	return err
}

func (c *cache) delAllBanned(uid string) error {
	keys, err := c.scanKeys(fmt.Sprintf("%s:b_%s-*", prefix, uid))
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	_, err = c.c.Del(keys...).Result()
	return err
}

func (c *cache) isBanned(uid string, rid int) (bool, error) {
	aff, err := c.c.Exists(keyBanned(uid, rid), keyBanned(uid, 0)).Result()
	return aff == 1, err
}

func (c *cache) scanKeys(key string) ([]string, error) {
	var keys []string
	var cur uint64
	var err error

	for {
		var tmpKeys []string
		tmpKeys, cur, err = c.c.Scan(cur, key, 100).Result()
		if err != nil {
			return nil, err
		}

		if len(tmpKeys) == 0 {
			break
		}

		keys = append(keys, tmpKeys...)

		if cur == 0 {
			break
		}
	}

	return keys, nil
}
