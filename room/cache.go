package room

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/zhenjl/cityhash"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	// _ "net/http/pprof"
)

type Cache interface {
	set(room models.Room) error
	get(id int) (models.Room, error)
	getChat(id int) (models.Room, error)
	setChatTopMessage(rids []int32, message []byte) error
	setChatBulletinMessage(rids []int32, message []byte) error
	getChatTopMessage(rid int) ([]byte, error)
	deleteChatTopMessage(rids []int32) error
	deleteChatBulletinMessage(rids []int32) error
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
	prefix = "ala"

	// 房間的前綴詞，用於存儲在redis當key
	roomKey = prefix + ":room_%d"

	roomDataHKey        = "data"
	roomTopMsgHKey      = "top"
	roomBulletinMsgHKey = "bulletin"

	// server name的前綴詞，用於存儲在redis當key
	onlineKey = prefix + ":server_%s"
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

func (c *cache) get(id int) (models.Room, error) {
	b, err := c.c.HGet(keyRoom(id), roomDataHKey).Bytes()
	if err != nil {
		return models.Room{}, err
	}
	var r models.Room
	if err := json.Unmarshal(b, &r); err != nil {
		return models.Room{}, err
	}
	return r, nil
}

func (c *cache) set(room models.Room) error {
	key := keyRoom(room.Id)

	tx := c.c.Pipeline()

	data := map[string]interface{}{}

	if room.TopMessage != nil {
		data[roomTopMsgHKey] = room.TopMessage
	}

	if room.BulletinMessage != nil {
		data[roomBulletinMsgHKey] = room.BulletinMessage
	}

	b1, err := json.Marshal(room)
	if err != nil {
		return err
	}

	data[roomDataHKey] = b1

	tx.HMSet(key, data)
	tx.Expire(key, roomExpired)
	_, err = tx.Exec()

	return err
}

func (c *cache) getChat(id int) (models.Room, error) {
	room, err := c.c.HMGet(keyRoom(id), roomDataHKey, roomTopMsgHKey, roomBulletinMsgHKey).Result()
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
	if room[1] != nil {
		r.TopMessage = []byte(room[1].(string))
	}

	if room[2] != nil {
		r.BulletinMessage = []byte(room[2].(string))
	}

	return r, err
}

func (c *cache) setChatTopMessage(rids []int32, message []byte) error {
	keys := make([]string, 0, len(rids))
	for _, rid := range rids {
		keys = append(keys, keyRoom(int(rid)))
	}

	tx := c.c.Pipeline()
	for _, k := range keys {
		tx.HSet(k, roomTopMsgHKey, message)
	}
	_, err := tx.Exec()
	return err
}

func (c *cache) setChatBulletinMessage(rids []int32, message []byte) error {
	keys := make([]string, 0, len(rids))
	for _, rid := range rids {
		keys = append(keys, keyRoom(int(rid)))
	}

	tx := c.c.Pipeline()
	for _, k := range keys {
		tx.HSet(k, roomBulletinMsgHKey, message)
	}
	_, err := tx.Exec()
	return err
}

func (c *cache) getChatTopMessage(rid int) ([]byte, error) {
	b, err := c.c.HGet(keyRoom(rid), roomTopMsgHKey).Bytes()
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, redis.Nil
	}
	return b, nil
}

func (c *cache) deleteChatTopMessage(rids []int32) error {
	keys := make([]string, 0, len(rids))
	for _, rid := range rids {
		keys = append(keys, keyRoom(int(rid)))
	}

	tx := c.c.Pipeline()
	for _, k := range keys {
		tx.HDel(k, roomTopMsgHKey)
	}
	_, err := tx.Exec()
	return err
}

func (c *cache) deleteChatBulletinMessage(rids []int32) error {
	keys := make([]string, 0, len(rids))
	for _, rid := range rids {
		keys = append(keys, keyRoom(int(rid)))
	}

	tx := c.c.Pipeline()
	for _, k := range keys {
		tx.HDel(k, roomBulletinMsgHKey)
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
		r := strconv.Itoa(int(room))
		rMap := roomsMap[cityhash.CityHash32([]byte(r), uint32(len(r)))%8]
		if rMap == nil {
			rMap = make(map[int32]int32)
			roomsMap[cityhash.CityHash32([]byte(r), uint32(len(r)))%8] = rMap
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
