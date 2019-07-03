package cache

import (
	"fmt"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
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

func (c *Cache) GetRoomByMoney(id string) (day, dml, deposit int, err error) {
	r, err := c.c.HMGet(keyRoom(id), hashLimitDayKey, hashLimitDmlKey, hashLimitAmountKey).Result()
	if err != nil {
		return 0, 0, 0, err
	}
	for _, v := range r {
		if v == nil {
			return 0, 0, 0, fmt.Errorf("room: %s limit data is nil", id)
		}
	}
	if day, err = strconv.Atoi(r[0].(string)); err != nil {
		return 0, 0, 0, fmt.Errorf("cache day limit: %s error(%v)", r[0], err)
	}
	if dml, err = strconv.Atoi(r[1].(string)); err != nil {
		return 0, 0, 0, fmt.Errorf("cache dml limit: %s error(%v)", r[1], err)
	}
	if deposit, err = strconv.Atoi(r[2].(string)); err != nil {
		return 0, 0, 0, fmt.Errorf("cache deposit limit: %s error(%v)", r[2], err)
	}
	return day, dml, deposit, nil
}

func (c *Cache) SetRoom(room models.Room) error {
	f := map[string]interface{}{
		hashPermissionKey:  room.Status(),
		hashLimitDayKey:    room.DayLimit,
		hashLimitDmlKey:    room.DmlLimit,
		hashLimitAmountKey: room.DepositLimit,
	}
	key := keyRoom(room.Id)
	tx := c.c.Pipeline()
	tx.HMSet(key, f)
	tx.Expire(key, roomExpired)
	_, err := tx.Exec()
	return err
}
