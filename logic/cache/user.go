package cache

import (
	"fmt"
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

func (c *Cache) GetUser(uid string, key string) (roomId, name string, status int, err error) {
	res, err := c.c.HMGet(keyUidInfo(uid), key, hashNameKey, hashStatusKey).Result()
	if err != nil {
		return "", "", 0, err
	}
	if len(res) != 3 {
		return "", "", 0, fmt.Errorf("conn.Receive() len is %d insufficient 3", len(res))
	}
	status, err = strconv.Atoi(res[2].(string))
	return res[0].(string), res[1].(string), status, err
}

// 更換房間
func (c *Cache) ChangeRoom(uid, key, roomId string) error {
	tx := c.c.Pipeline()
	tx.HSet(keyUidInfo(uid), key, roomId)
	tx.Expire(keyUidInfo(uid), c.expire)
	_, err := tx.Exec()
	return err
}
