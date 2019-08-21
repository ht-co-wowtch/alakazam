package member

import (
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

func keyUidInfo(uid string) string {
	return fmt.Sprintf(uidKey, uid)
}

func keyBannedInfo(uid string) string {
	return fmt.Sprintf(bannedKey, uid)
}

const (
	// user hash table name key
	hashNameKey = "name"

	// user hash table status key
	hashStatusKey = "status"

	// user hash table server key
	hashServerKey = "server"

	hMidKey = "mid"
)

// 儲存user資訊
func (c *Cache) set(member *models.Member, key, roomId, server string) error {
	keyI := keyUidInfo(member.Uid)
	tx := c.c.Pipeline()
	f := map[string]interface{}{
		key:           roomId,
		hMidKey:       member.Id,
		hashNameKey:   member.Name,
		hashStatusKey: member.Status(),
		hashServerKey: server,
	}
	tx.HMSet(keyI, f)
	tx.Expire(keyI, c.expire)
	_, err := tx.Exec()
	return err
}

var errUserNil = errors.New("get user cache data has nil")

type HMember struct {
	Mid    int
	Room   int
	Name   string
	Status int
}

func (c *Cache) get(uid string, key string) (HMember, error) {
	res, err := c.c.HMGet(keyUidInfo(uid), key, hashNameKey, hashStatusKey, hMidKey).Result()
	if err != nil {
		return HMember{}, err
	}
	for _, v := range res {
		if v == nil {
			return HMember{}, errdefs.InvalidParameter(errUserNil, 1)
		}
	}
	rid, err := strconv.Atoi(res[0].(string))
	if err != nil {
		return HMember{}, err
	}
	status, err := strconv.Atoi(res[2].(string))
	if err != nil {
		return HMember{}, err
	}
	mid, err := strconv.Atoi(res[3].(string))
	if err != nil {
		return HMember{}, err
	}
	return HMember{
		Mid:    mid,
		Room:   rid,
		Name:   res[1].(string),
		Status: status,
	}, nil
}

// 設定禁言
func (c *Cache) setBanned(uid string, expired time.Duration) error {
	key := keyBannedInfo(uid)
	i, err := c.c.Exists(key).Result()
	if err != nil {
		return err
	}
	tx := c.c.Pipeline()
	tx.Set(key, time.Now().Add(expired).Unix(), expired)
	if i <= 0 {
		tx.HIncrBy(keyUidInfo(uid), hashStatusKey, -models.MessageStatus)
	}
	_, err = tx.Exec()
	return err
}

// 取得禁言時效
func (c *Cache) getBanned(uid string) (time.Time, bool, error) {
	sec, err := c.c.Get(keyBannedInfo(uid)).Int64()
	if err != nil {
		if err == redis.Nil {
			err = nil
		}
		return time.Time{}, false, err
	}
	return time.Unix(sec, 0), true, nil
}

// 解除禁言
func (c *Cache) delBanned(uid string) error {
	tx := c.c.Pipeline()
	tx.Del(keyBannedInfo(uid))
	tx.HIncrBy(keyUidInfo(uid), hashStatusKey, models.MessageStatus)
	_, err := tx.Exec()
	return err
}

// restart user資料的過期時間
// EXPIRE : uid_{user id}  (HSET)
func (c *Cache) refreshUserExpire(uid string) (bool, error) {
	return c.c.Expire(keyUidInfo(uid), c.expire).Result()
}

// 移除user資訊
// DEL : uid_{user id}
func (c *Cache) deleteUser(uid, key string) (bool, error) {
	aff, err := c.c.HDel(keyUidInfo(uid), key).Result()
	return aff >= 1, err
}

// 取會員名稱
// TODO user name 資料結構需要優化，不然這樣 redis O(n)
func (c *Cache) getName(uid []string) ([]string, error) {
	tx := c.c.Pipeline()
	cmd := make([]*redis.StringCmd, len(uid))
	for i, id := range uid {
		cmd[i] = tx.HGet(keyUidInfo(id), hashNameKey)
	}
	_, err := tx.Exec()
	if err != nil {
		return nil, err
	}
	name := make([]string, len(uid))
	for i, v := range cmd {
		name[i] = v.Val()
	}
	return name, nil
}

// 更換房間
func (c *Cache) changeRoom(uid, key, roomId string) error {
	tx := c.c.Pipeline()
	tx.HSet(keyUidInfo(uid), key, roomId)
	tx.Expire(keyUidInfo(uid), c.expire)
	_, err := tx.Exec()
	return err
}
