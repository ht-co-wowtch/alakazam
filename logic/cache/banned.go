package cache

import (
	"github.com/go-redis/redis"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"time"
)

// 設定禁言
func (c *Cache) SetBanned(uid string, expired time.Duration) error {
	key := keyBannedInfo(uid)
	i, err := c.c.Exists(key).Result()
	if err != nil {
		return err
	}
	tx := c.c.Pipeline()
	tx.Set(key, time.Now().Add(expired).Unix(), expired)
	if i <= 0 {
		tx.HIncrBy(keyUidInfo(uid), hashStatusKey, -models.Message)
	}
	_, err = tx.Exec()
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
	tx.HIncrBy(keyUidInfo(uid), hashStatusKey, models.Message)
	_, err := tx.Exec()
	return err
}
