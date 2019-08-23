package room

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/zhenjl/cityhash"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"strconv"
	"time"
)

type Cache struct {
	c *redis.Client
}

func newCache(client *redis.Client) *Cache {
	return &Cache{
		c: client,
	}
}

const (
	// 房間的前綴詞，用於存儲在redis當key
	roomKey = "room_%s"

	userRoomKey = "uid_room_%s"

	// server name的前綴詞，用於存儲在redis當key
	onlineKey = "server_%s"
)

func keyRoom(id string) string {
	return fmt.Sprintf(roomKey, id)
}

func keyUserRoom(id string) string {
	return fmt.Sprintf(userRoomKey, id)
}

func keyServerOnline(key string) string {
	return fmt.Sprintf(onlineKey, key)
}

var (
	roomExpired = time.Hour
)

func (c *Cache) set(room models.Room) error {
	b, err := json.Marshal(room)
	if err != nil {
		return err
	}
	return c.c.Set(keyRoom(strconv.Itoa(room.Id)), b, roomExpired).Err()
}

func (c *Cache) get(id string) (*models.Room, error) {
	b, err := c.c.Get(keyRoom(id)).Bytes()
	if err != nil {
		return nil, err
	}
	var r models.Room
	err = json.Unmarshal(b, &r)
	return &r, err
}

type Online struct {
	Server    string           `json:"server"`
	RoomCount map[string]int32 `json:"room_count"`
	Updated   int64            `json:"updated"`
}

// 以HSET方式儲存房間人數
// HSET Key hashKey jsonBody
// Key用server name
// hashKey則是將room name以City Hash32做hash後得出一個數字，以這個數字當hashKey
// TODO 需要在思考是否需要這樣的機制
func (c *Cache) addOnline(server string, online *Online) error {
	roomsMap := map[uint32]map[string]int32{}
	for room, count := range online.RoomCount {
		rMap := roomsMap[cityhash.CityHash32([]byte(room), uint32(len(room)))%8]
		if rMap == nil {
			rMap = make(map[string]int32)
			roomsMap[cityhash.CityHash32([]byte(room), uint32(len(room)))%8] = rMap
		}
		rMap[room] = count
	}

	key := keyServerOnline(server)
	for hashKey, value := range roomsMap {
		err := c.addServerOnline(key, strconv.FormatInt(int64(hashKey), 10), &Online{RoomCount: value, Server: online.Server, Updated: online.Updated})
		if err != nil {
			return err
		}
	}
	return nil
}

// 以HSET方式儲存房間人數
// HSET Key hashKey jsonBody
// Key用server name
func (c *Cache) addServerOnline(key string, hashKey string, online *Online) error {
	b, err := json.Marshal(online)
	if err != nil {
		return err
	}
	tx := c.c.Pipeline()
	tx.HSet(key, hashKey, b)
	tx.Expire(key, roomExpired)
	_, err = tx.Exec()
	return err
}

// 根據server name取線上各房間總人數
// TODO 需要在思考需不需要比對Updated
func (c *Cache) getOnline(server string) (*Online, error) {
	online := &Online{RoomCount: map[string]int32{}}
	// server name
	key := keyServerOnline(server)
	for i := 0; i < 8; i++ {
		ol, err := c.serverOnline(key, strconv.FormatInt(int64(i), 10))
		if err == nil && ol != nil {
			online.Server = ol.Server
			if ol.Updated > online.Updated {
				online.Updated = ol.Updated
			}
			for room, count := range ol.RoomCount {
				online.RoomCount[room] = count
			}
		}
	}
	return online, nil
}

// 根據server name與hashKey取該server name內線上各房間總人數
func (c *Cache) serverOnline(key string, hashKey string) (*Online, error) {
	b, err := c.c.HGet(key, hashKey).Bytes()
	if err != nil && err != redis.Nil {
		return nil, err
	}
	// b是一個json
	// {
	// 		"server":"ne0002de-MacBook-Pro.local",
	// 		"room_count":{
	// 			"1000":1
	// 		 },
	// 		 "updated":1556368160
	// }"
	// 1000是房間id，1是人數
	// updated是資料更新時間
	online := new(Online)
	if err = json.Unmarshal(b, online); err != nil {
		return nil, err
	}
	return online, nil
}

// 根據server name 刪除線上各房間總人數
func (c *Cache) delOnline(server string) error {
	return c.c.Del(keyServerOnline(server)).Err()
}
