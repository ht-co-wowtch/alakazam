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

type Cache interface {
	set(room models.Room) error
	get(id int) (models.Room, error)
	setChat(room models.Room, message string) error
	getChat(id int) (models.Room, error)
	setChatTopMessage(rids []int32, message string) error
	deleteChatTopMessage(rids []int32) error
	addOnline(server string, online *Online) error
	getOnline(server string) (*Online, error)
}

type cache struct {
	c *redis.Client
}

func newCache(client *redis.Client) Cache {
	return &cache{
		c: client,
	}
}

const (
	// 房間的前綴詞，用於存儲在redis當key
	roomKey = "room_%d"

	roomDataKey   = "data"
	roomTopMsgKey = "msg"

	// server name的前綴詞，用於存儲在redis當key
	onlineKey = "server_%s"
)

func keyRoom(id int) string {
	return fmt.Sprintf(roomKey, id)
}

func keyServerOnline(key string) string {
	return fmt.Sprintf(onlineKey, key)
}

var (
	roomExpired = time.Hour * 12
)

func (c *cache) set(room models.Room) error {
	tx := c.c.Pipeline()
	b, err := json.Marshal(room)
	if err != nil {
		return err
	}
	key := keyRoom(room.Id)
	tx.HSet(key, roomDataKey, b)
	tx.Expire(key, roomExpired)
	_, err = tx.Exec()
	return err
}

func (c *cache) get(id int) (models.Room, error) {
	b, err := c.c.HGet(keyRoom(id), roomDataKey).Bytes()
	if err != nil {
		return models.Room{}, err
	}
	var r models.Room
	if err := json.Unmarshal(b, &r); err != nil {
		return models.Room{}, err
	}
	return r, nil
}

func (c *cache) setChat(room models.Room, message string) error {
	b1, err := json.Marshal(room)
	if err != nil {
		return err
	}

	tx := c.c.Pipeline()
	key := keyRoom(room.Id)
	tx.HMSet(key, map[string]interface{}{
		roomDataKey:   b1,
		roomTopMsgKey: message,
	})
	tx.Expire(key, roomExpired)
	_, err = tx.Exec()
	return err
}

func (c *cache) getChat(id int) (models.Room, error) {
	room, err := c.c.HMGet(keyRoom(id), roomDataKey, roomTopMsgKey).Result()
	if err != nil {
		return models.Room{}, err
	}
	if room[0] == nil {
		return models.Room{}, redis.Nil
	}

	var r models.Room
	if err = json.Unmarshal([]byte(room[0].(string)), &r); err != nil {
		return r, err
	}
	if room[1] == nil {
		return r, nil
	}

	r.TopMessage = room[1].(string)
	return r, err
}

func (c *cache) setChatTopMessage(rids []int32, message string) error {
	keys := make([]string, 0, len(rids))
	for _, rid := range rids {
		keys = append(keys, keyRoom(int(rid)))
	}

	tx := c.c.Pipeline()
	for _, k := range keys {
		tx.HSet(k, roomTopMsgKey, message)
	}
	_, err := tx.Exec()
	return err
}

func (c *cache) deleteChatTopMessage(rids []int32) error {
	keys := make([]string, 0, len(rids))
	for _, rid := range rids {
		keys = append(keys, keyRoom(int(rid)))
	}

	tx := c.c.Pipeline()
	for _, k := range keys {
		tx.HDel(k, roomTopMsgKey)
	}
	_, err := tx.Exec()
	return err
}

type Online struct {
	Server    string          `json:"server"`
	RoomCount map[int32]int32 `json:"room_count"`
	Updated   int64           `json:"updated"`
}

// 以HSET方式儲存房間人數
// HSET Key hashKey jsonBody
// Key用server name
// hashKey則是將room name以City Hash32做hash後得出一個數字，以這個數字當hashKey
// TODO 需要在思考是否需要這樣的機制
func (c *cache) addOnline(server string, online *Online) error {
	roomsMap := map[uint32]map[int32]int32{}
	for room, count := range online.RoomCount {
		rMap := roomsMap[cityhash.CityHash32([]byte(strconv.Itoa(int(room))), uint32(room))%8]
		if rMap == nil {
			rMap = make(map[int32]int32)
			roomsMap[cityhash.CityHash32([]byte(strconv.Itoa(int(room))), uint32(room))%8] = rMap
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
func (c *cache) addServerOnline(key string, hashKey string, online *Online) error {
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
func (c *cache) getOnline(server string) (*Online, error) {
	online := &Online{RoomCount: map[int32]int32{}}
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
func (c *cache) serverOnline(key string, hashKey string) (*Online, error) {
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
func (c *cache) delOnline(server string) error {
	return c.c.Del(keyServerOnline(server)).Err()
}
