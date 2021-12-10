package room

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/zhenjl/cityhash"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	// _ "runtime/pprof"
)

//
// 快取操作
//

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
	addOnlineViewer(server string, viewers *OnlineViewer) error
	getOnlineViewer(server string) (*OnlineViewer, error)
	addPayment(uid string, liveChatId int, paidTime time.Time, addPayment float32) error
	getPayment(uid string, liveChatId int) (*Payment, error)
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

	roomDataHKey = "data"

	roomTopMsgHKey      = "top"
	roomBulletinMsgHKey = "bulletin"

	// server name的前綴詞，用於存儲在redis當key
	onlineKey = prefix + ":server_%s"

	onlineViewerKey = prefix + ":server_%s_viewer"

	uidPayKey = prefix + ":pay_%s-%d"

	livePaymentDataHKey = "livePayment"
)

func keyRoom(id int) string {
	return fmt.Sprintf(roomKey, id)
}

func keyServerOnline(key string) string {
	return fmt.Sprintf(onlineKey, key)
}

func keyServerOnlineViewer(key string) string {
	return fmt.Sprintf(onlineViewerKey, key)
}

func keyPaid(uid string, rid int) string {
	return fmt.Sprintf(uidPayKey, uid, rid)
}

var (
	roomExpired = time.Hour * 12
	payExpired  = time.Second * 55 // 根據前端需求改修正為小於一分鐘(60秒)
)

// 取得房間快取
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

// 設置房間的快取
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

// 從快取中取得聊天室資訊
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

// 設置置頂訊息快取
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

// 設置公告訊息快取
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

// 從快取中取得置頂訊息
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

// 刪除置頂訊息快取
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

// 刪除公告訊息快取
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

func (c *cache) _useDefault(rids []int32) error {
	//testing only
	id := int(rids[0])

	room, err := c.c.HMGet(keyRoom(id), roomDataHKey, roomTopMsgHKey, roomBulletinMsgHKey).Result()

	if err != nil {
		return err
	}
	if room[0] == nil {
		return redis.Nil
	}

	var r models.Room
	if err = json.Unmarshal([]byte(room[0].(string)), &r); err != nil {
		return err
	}
	if room[1] != nil {
		r.TopMessage = []byte(room[1].(string))
	}

	if room[2] != nil {
		r.BulletinMessage = []byte(room[2].(string))
	}

	return err
}

type Online struct {
	Server    string          `json:"server"`
	RoomCount map[int32]int32 `json:"room_count"`
	Updated   int64           `json:"updated"`
}

type OnlineViewer struct {
	Server      string             `json:"server"`
	RoomViewers map[int32][]string `json:"room_viewers"`
	Updated     int64              `json:"updated"`
}

type Payment struct {
	PaidTime string  `json:"paid_time"`
	Diamond  float32 `json:"diamond"`
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

// 以HSET方式儲存房間觀眾
func (c *cache) addOnlineViewer(server string, onlineViewer *OnlineViewer) error {
	roomsMap := map[uint32]map[int32][]string{}
	for room, viewers := range onlineViewer.RoomViewers {
		r := strconv.Itoa(int(room))
		rMap := roomsMap[cityhash.CityHash32([]byte(r), uint32(len(r)))%8]
		if rMap == nil {
			rMap = make(map[int32][]string)
			roomsMap[cityhash.CityHash32([]byte(r), uint32(len(r)))%8] = rMap
		}
		rMap[room] = viewers
	}

	key := keyServerOnlineViewer(server)
	for hashKey, value := range roomsMap {
		err := c.addServerOnlineViewers(
			key,
			strconv.FormatInt(int64(hashKey), 10),
			&OnlineViewer{
				RoomViewers: value,
				Server:      onlineViewer.Server,
				Updated:     onlineViewer.Updated,
			},
		)
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

// 以HSET方式儲存房間人數UID
// HSET Key hashKey jsonBody
// Key用server name
func (c *cache) addServerOnlineViewers(key string, hashKey string, viewers *OnlineViewer) error {
	b, err := json.Marshal(viewers)
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

// 根據server name取線上各房間所有人UID
// TODO
func (c *cache) getOnlineViewer(server string) (*OnlineViewer, error) {
	viewerOnline := &OnlineViewer{RoomViewers: map[int32][]string{}}
	// server name
	key := keyServerOnlineViewer(server)
	for i := 0; i < 8; i++ {
		olw, err := c.serverOnlineViewer(key, strconv.FormatInt(int64(i), 10))
		if err == nil && olw != nil {
			viewerOnline.Server = olw.Server
			if olw.Updated > viewerOnline.Updated {
				viewerOnline.Updated = olw.Updated
			}
			for room, viewers := range olw.RoomViewers {
				viewerOnline.RoomViewers[room] = viewers
			}

			viewerOnline.Updated = olw.Updated
		}
	}

	return viewerOnline, nil
}

// 根據server name與hashKey取該server name內線上各房間總人數
func (c *cache) serverOnline(key string, hashKey string) (*Online, error) {
	b, err := c.c.HGet(key, hashKey).Bytes()
	if err != nil && err != redis.Nil {
		return nil, err
	}
	//log.Infof("serverOnline data:%+v", string(b))
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

// 根據server name與hashKey取該server name內線上各房間所有人UID
func (c *cache) serverOnlineViewer(key string, hashKey string) (*OnlineViewer, error) {
	b, err := c.c.HGet(key, hashKey).Bytes()
	if err != nil && err != redis.Nil {
		return nil, err
	}

	//onlineViewer := new(OnlineViewer)
	viewer := new(OnlineViewer)
	if err = json.Unmarshal(b, viewer); err != nil {
		return nil, err
	}

	return viewer, nil
}

// 根據server name 刪除線上各房間總人數
func (c *cache) delOnline(server string) error {
	return c.c.Del(keyServerOnline(server)).Err()
}

func (c *cache) addPayment(uid string, liveChatId int, paidTime time.Time, diamond float32) error {
	key := keyPaid(uid, liveChatId)

	data := map[string]interface{}{}

	b1, err := json.Marshal(Payment{
		PaidTime: paidTime.Format(time.RFC3339),
		Diamond:  diamond,
	})

	if err != nil {
		return err
	}

	data[livePaymentDataHKey] = b1

	tx := c.c.Pipeline()
	tx.HMSet(key, data)
	tx.Expire(key, payExpired)
	_, err = tx.Exec()

	return err
}

func (c *cache) getPayment(uid string, liveChatId int) (*Payment, error) {
	key := keyPaid(uid, liveChatId)

	b, err := c.c.HGet(key, livePaymentDataHKey).Bytes()
	if err != nil && err != redis.Nil {
		return nil, err
	}

	p := new(Payment)
	if err = json.Unmarshal(b, p); err != nil {
		return nil, err
	}

	return p, nil
}
