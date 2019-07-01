package cache

import (
	"errors"
	"github.com/go-redis/redis"
	"strconv"
)

// 儲存user資訊
// HSET :
// 主key => uid_{user id}
// user key => user roomId
// name => user name
// status => user status
// token => 三方應用接口token
// server => comet server name
func (c *Cache) SetUser(uid, key, roomId, name, server string, status int) error {
	keyI := keyUidInfo(uid)
	tx := c.c.Pipeline()
	f := map[string]interface{}{
		key:           roomId,
		hashNameKey:   name,
		hashStatusKey: status,
		hashServerKey: server,
	}
	tx.HMSet(keyI, f)
	tx.Expire(keyI, c.expire)
	_, err := tx.Exec()
	return err
}

// restart user資料的過期時間
// EXPIRE : uid_{user id}  (HSET)
func (c *Cache) RefreshUserExpire(uid string) (bool, error) {
	return c.c.Expire(keyUidInfo(uid), c.expire).Result()
}

// 移除user資訊
// DEL : uid_{user id}
func (c *Cache) DeleteUser(uid, key string) (bool, error) {
	aff, err := c.c.HDel(keyUidInfo(uid), key).Result()
	return aff >= 1, err
}

var errUserNil = errors.New("get user cache data has nil")

func (c *Cache) GetUser(uid string, key string) (roomId, name string, status int, err error) {
	res, err := c.c.HMGet(keyUidInfo(uid), key, hashNameKey, hashStatusKey).Result()
	if err != nil {
		return "", "", 0, err
	}
	for _, v := range res {
		if v == nil {
			return "", "", 0, errUserNil
		}
	}
	if status, err = strconv.Atoi(res[2].(string)); err != nil {
		return "", "", 0, err
	}
	return res[0].(string), res[1].(string), status, err
}

// 取會員名稱
// TODO user name 資料結構需要優化，不然這樣 redis O(n)
func (c *Cache) GetUserName(uid []string) ([]string, error) {
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
func (c *Cache) ChangeRoom(uid, key, roomId string) error {
	tx := c.c.Pipeline()
	tx.HSet(keyUidInfo(uid), key, roomId)
	tx.Expire(keyUidInfo(uid), c.expire)
	_, err := tx.Exec()
	return err
}
