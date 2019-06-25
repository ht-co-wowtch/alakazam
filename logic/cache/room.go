package cache

import (
	"strconv"
	"time"
)

const (
	hashPermissionKey = "permission"

	hashLimitDayKey = "day"

	hashLimitAmountKey = "amount"

	hashLimitDmlKey = "dml"
)

var (
	roomExpired = time.Hour
)

func (c *Cache) GetRoom(id string) (int, error) {
	return c.c.HGet(keyRoom(id), hashPermissionKey).Int()
}

func (c *Cache) GetRoomByMoney(id string) (day, dml, amount int, err error) {
	r, err := c.c.HMGet(keyRoom(id), hashLimitDayKey, hashLimitDmlKey, hashLimitAmountKey).Result()
	if err != nil {
		return 0, 0, 0, err
	}
	i := make([]int, 3)
	for k, _ := range i {
		i[k], _ = strconv.Atoi(r[k].(string))
	}
	return i[0], i[1], i[2], err
}

func (c *Cache) SetRoom(id string, permission, day, dml, amount int) error {
	f := map[string]interface{}{
		hashPermissionKey:  permission,
		hashLimitDayKey:    day,
		hashLimitDmlKey:    dml,
		hashLimitAmountKey: amount,
	}
	key := keyRoom(id)
	tx := c.c.Pipeline()
	tx.HMSet(key, f)
	tx.Expire(key, roomExpired)
	_, err := tx.Exec()
	return err
}
