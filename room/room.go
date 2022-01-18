package room

import (
	"database/sql"
	"encoding/json"
	"time"

	"gitlab.com/ht-co/micro/log"

	"github.com/go-redis/redis"
	"github.com/go-sql-driver/mysql"
	"gitlab.com/ht-co/wowtch/live/alakazam/client"
	"gitlab.com/ht-co/wowtch/live/alakazam/errors"
	"gitlab.com/ht-co/wowtch/live/alakazam/member"
	"gitlab.com/ht-co/wowtch/live/alakazam/message"
	"gitlab.com/ht-co/wowtch/live/alakazam/message/scheme"
	"gitlab.com/ht-co/wowtch/live/alakazam/models"
)

type Room interface {
	Create(status Status) (int, error)
	Update(id int, status Status) error
	Delete(id int) error
	Get(id int) (models.Room, error)
	GetTopMessage(msgId int64, t int) ([]int32, models.Message, error)
	AddTopMessage(rids []int32, seq int64, message string, ts []int) error
	DeleteTopMessage(rids []int32, msgId int64, t int) error
	Online() (map[int32]int32, error)
	OnlineViewer() (map[int32][]string, error)
}

type room struct {
	db     *models.Store
	c      Cache
	member *member.Member
	cli    *client.Client
}

func New(db *models.Store, cache *redis.Client, m *member.Member, cli *client.Client) Room {
	return &room{
		db:     db,
		c:      newCache(cache),
		member: m,
		cli:    cli,
	}
}

type Status struct {
	// 是否禁言
	IsMessage bool `json:"is_message"`

	// 是否打開跟注
	IsBets bool `json:"is_bets"`

	// 儲值&打碼量發話限制
	Limit Limit `json:"limit"`

	// 觀眾數倍率
	AudienceRatio float64 `json:"audience_ratio"`

	// 房間狀態
	Status bool `json:"status"`
}

type Limit struct {
	// 限制範圍
	Day int `json:"day" binding:"max=31"`

	// 儲值金額
	Deposit int `json:"deposit"`

	// 打碼量
	Dml int `json:"dml"`
}

// Create
// 新增房間
func (r *room) Create(status Status) (int, error) {
	model := r.newRoomModel(status)
	model.Status = true

	if _, err := r.db.CreateRoom(&model); err != nil {
		return 0, err
	}
	return model.Id, r.c.set(model)
}

// Update
// 更新房間
func (r *room) Update(id int, status Status) error {
	model := r.newRoomModel(status)
	model.Id = id
	model.Status = status.Status

	if _, err := r.db.UpdateRoom(model); err != nil {
		return err
	}
	if err := r.c.set(model); err != nil {
		return err
	}
	return nil
}

// 建立房間room Model
func (r *room) newRoomModel(status Status) models.Room {
	if status.AudienceRatio < 1 {
		status.AudienceRatio = 1
	}

	return models.Room{
		IsMessage:     status.IsMessage,
		IsBets:        status.IsBets,
		DayLimit:      status.Limit.Day,
		DepositLimit:  status.Limit.Deposit,
		DmlLimit:      status.Limit.Dml,
		AudienceRatio: status.AudienceRatio,
	}
}

// Delete
// 刪除房間
func (r *room) Delete(id int) error {
	room, err := r.Get(id)
	if err != nil {
		return err
	}
	if room.Status == false {
		return nil
	}
	aff, err := r.db.DeleteRoom(id)
	if err != nil {
		return err
	}
	if aff <= 0 {
		return errors.ErrNoRows
	}

	room.Status = false
	if err := r.c.set(room); err != nil {
		return err
	}
	return nil
}

// Get
// 取得房間資料
func (r *room) Get(id int) (models.Room, error) {

	room, err := r.db.GetRoom(id)

	if err != nil {
		if err == sql.ErrNoRows {
			return models.Room{}, errors.ErrNoRows
		}
		return models.Room{}, err
	}
	return room, nil
}

// GetTopMessage
// 取得置頂/公告訊息 By msg_id & type id
func (r *room) GetTopMessage(msgId int64, t int) ([]int32, models.Message, error) {
	roomTopMsg, err := r.db.FindRoomTopMessage(msgId) // 從db中取得訊息 by msg_id
	if err != nil {
		return nil, models.Message{}, err
	}
	if len(roomTopMsg) == 0 {
		return nil, models.Message{}, errors.ErrNoRows
	}

	var index int
	rid := make([]int32, 0, len(roomTopMsg))
	for i, v := range roomTopMsg {
		if v.Type == t {
			index = i
			rid = append(rid, v.RoomId)
		}
	}

	return rid, models.Message{
		MsgId:   msgId,
		Message: roomTopMsg[index].Message,
		SendAt:  roomTopMsg[index].SendAt,
	}, nil
}

// AddTopMessage
// 新增置頂/公告訊息
func (r *room) AddTopMessage(rids []int32, seq int64, msg string, ts []int) error {
	var roomTopMessage scheme.Message
	model := models.RoomTopMessage{
		MsgId:   seq,
		Message: msg,
		SendAt:  time.Now(),
	}

	for _, t := range ts {
		model.Type = t
		if err := r.db.AddRoomTopMessage(rids, model); err != nil {
			if e, ok := err.(*mysql.MySQLError); ok {
				if e.Number == 1452 {
					return errors.ErrNoRoom
				}
			}
			return err
		}

		if t == models.TOP_MESSAGE {
			roomTopMessage = message.RoomTopMessageToMessage(model) // 置頂訊息
		} else {
			roomTopMessage = message.RoomBulletinMessageToMessage(model) // 公告訊息
		}

		b, err := json.Marshal(roomTopMessage)
		if err != nil {
			return err
		}

		if t == models.TOP_MESSAGE {
			err = r.c.setChatTopMessage(rids, b) // 置頂訊息寫入到Cache
		} else {
			err = r.c.setChatBulletinMessage(rids, b) //公告訊息寫入到Cache
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteTopMessage
// 刪除置頂/公告訊息
func (r *room) DeleteTopMessage(rids []int32, msgId int64, t int) error {
	err := r.db.DeleteRoomTopMessage(rids, msgId, t)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.ErrNoRows
		}
		return err
	}

	if t == models.TOP_MESSAGE {
		return r.c.deleteChatTopMessage(rids) //從cache中刪除置頂訊息
	}

	return r.c.deleteChatBulletinMessage(rids) //從cache中刪除公告訊息
}

// Online
// 從快取中取得所有房間在線人數
func (r *room) Online() (map[int32]int32, error) {
	//底下的hostname會用於快取的key,與comet/server.go - NewServer - s.name = "hostname"
	online, err := r.c.getOnline("hostname")
	if err != nil {
		return nil, err
	}
	return online.RoomCount, nil
}

// OnlineViewer
// 從快取中取得所有房間在線人UID
func (r *room) OnlineViewer() (map[int32][]string, error) {
	viewer, err := r.c.getOnlineViewer("hostname")
	if err != nil {
		return nil, err
	}

	log.Infof("OnlineViewer, %+v", viewer)
	return viewer.RoomViewers, nil
}
