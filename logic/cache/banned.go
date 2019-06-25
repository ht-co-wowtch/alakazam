package cache

import (
	"github.com/go-redis/redis"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/permission"
	"time"
)

// TODO expired 參數要改成time.Duration
// 設定禁言
func (c *Cache) SetBanned(uid string, expired int) error {
	sec := time.Duration(expired) * time.Second
	tx := c.c.Pipeline()
	tx.Set(keyBannedInfo(uid), time.Now().Add(sec).Unix(), sec)
	tx.HIncrBy(keyUidInfo(uid), hashStatusKey, -permission.Message)
	_, err := tx.Exec()
	return err
}

// 取得禁言時效
func (c *Cache) GetBanned(uid string) (time.Time, bool, error) {
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
func (c *Cache) DelBanned(uid string) error {
	tx := c.c.Pipeline()
	tx.Del(keyBannedInfo(uid))
	tx.HIncrBy(keyUidInfo(uid), hashStatusKey, permission.Message)
	_, err := tx.Exec()
	return err
}
