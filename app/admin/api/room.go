package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gitlab.com/jetfueltw/cpw/alakazam/message/scheme"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"gitlab.com/jetfueltw/cpw/alakazam/room"
	"gitlab.com/ht-co/cpw/micro/log"
	"go.uber.org/zap"
	//_ "net/http/pprof"
)

// 新增房間
func (s *httpServer) CreateRoom(c *gin.Context) error {
	var params room.Status
	if err := c.ShouldBindJSON(&params); err != nil {
		return err
	}
	roomId, err := s.room.Create(params)
	if err != nil {
		return err
	}
	c.JSON(http.StatusOK, gin.H{
		"id": roomId,
	})
	return nil
}

// 更新房間設定
func (s *httpServer) UpdateRoom(c *gin.Context) error {
	var params room.Status
	var r models.Room
	rid, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return err
	}
	if err := c.ShouldBindJSON(&params); err != nil {
		return err
	}
	// 從DB取得房間資料
	if r, err = s.room.Get(rid); err != nil {
		return err
	}
	// 更新房間設定
	if err := s.room.Update(rid, params); err != nil {
		return err
	}

	if len(s.noticeUrl) > 0 {
		for _, u := range s.noticeUrl {
			b, _ := json.Marshal(gin.H{
				"id":         r.Id,
				"is_message": r.IsMessage,
				"is_bets":    r.IsBets,
				"limit": room.Limit{
					Day:     r.DayLimit,
					Deposit: r.DepositLimit,
					Dml:     r.DmlLimit,
				},
				"status":    r.Status,
				"create_at": r.CreateAt,
				"update_at": r.UpdateAt,
			})

			if _, err := http.Post(u, "application/json", bytes.NewReader(b)); err != nil {
				log.Error("notice", zap.Int("rid", rid), zap.Error(err))
			}
		}
	}

	c.Status(http.StatusNoContent)
	return nil
}

type Rid struct {
	Id int `form:"id" binding:"required"`
}

// 取得房間
func (s *httpServer) GetRoom(c *gin.Context) error {
	rid, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return err
	}
	// 請求資料欄位驗證
	if err := binding.Validator.ValidateStruct(&Rid{Id: rid}); err != nil {
		return err
	}
	r, err := s.room.Get(rid) // 從DB取得房間資料
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         r.Id,
		"is_message": r.IsMessage,
		"is_bets":    r.IsBets,
		"limit": room.Limit{
			Day:     r.DayLimit,
			Deposit: r.DepositLimit,
			Dml:     r.DmlLimit,
		},
		"status":    r.Status,
		"create_at": r.CreateAt,
		"update_at": r.UpdateAt,
	})
	return nil
}

// 刪除房間
func (s *httpServer) DeleteRoom(c *gin.Context) error {
	rid, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return err
	}
	// 請求資料欄位驗證
	if err := binding.Validator.ValidateStruct(&Rid{Id: rid}); err != nil {
		return err
	}
	// 刪除房間
	if err := s.room.Delete(rid); err != nil {
		return err
	}
	c.Status(http.StatusNoContent)
	return nil
}

// 新增房間管理員
func (s *httpServer) AddManage(c *gin.Context) error {
	var params struct {
		RoomId int    `json:"room_id" binding:"required"`
		Uid    string `json:"uid" binding:"required,len=32"`
	}

	if err := c.ShouldBindJSON(&params); err != nil {
		return err
	}

	// DB中寫入會員於房間內權限
	if err := s.member.SetManage(params.Uid, params.RoomId, true); err != nil {
		return err
	}

	keys, err := s.member.GetRoomKeys(params.Uid, params.RoomId) // 從快去中取得會員與房間ws連線key值
	if err != nil {
		return err
	}

	m, _ := s.member.GetByRoom(params.Uid, params.RoomId) // 從快取中取得會員在該房間權限
	r, _ := s.room.Get(params.RoomId) // 從DB取得房間資料
	connect := room.NewPbConnect(m, r, "", 0)

	if len(keys) > 0 {
		_, _ = s.message.SendPermission(keys, m, *connect)
		_, _ = s.message.SendDisplay([]int32{int32(params.RoomId)}, scheme.NewRoot(), scheme.DisplayBySetManage(m.Name, true))
	}

	c.Status(http.StatusNoContent)
	return nil
}

// 刪除房間管理員
func (s *httpServer) DeleteManage(c *gin.Context) error {
	rid, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return err
	}

	params := struct {
		RoomId int    `json:"room_id" binding:"required"`
		Uid    string `form:"uid" binding:"required,len=32"`
	}{
		RoomId: rid,
		Uid:    c.Param("uid"),
	}
	// 請求資料欄位驗證
	if err := binding.Validator.ValidateStruct(&params); err != nil {
		return err
	}

	if err := s.member.SetManage(params.Uid, params.RoomId, false); err != nil {
		return err
	}
	keys, err := s.member.GetKeys(params.Uid) // 從快取中取得會員資料
	if err != nil {
		return err
	}
	m, _ := s.member.GetByRoom(params.Uid, params.RoomId) // 從快取中取得會員在該房間權限
	r, _ := s.room.Get(int(params.RoomId)) // 從DB取得房間資料
	connect := room.NewPbConnect(m, r, "", 0)

	_, _ = s.message.SendPermission(keys, m, *connect)
	_, _ = s.message.SendDisplay([]int32{int32(params.RoomId)}, scheme.NewRoot(), scheme.DisplayBySetManage(m.Name, false))

	c.Status(http.StatusNoContent)
	return nil
}

// 所有房間在線人數
func (s *httpServer) online(c *gin.Context) error {
	o, err := s.room.Online() // 從快取中取得所有房間在線人數
	if err == nil {
		c.JSON(http.StatusOK, o)
	}
	return err
}

func (s *httpServer) notice(c *gin.Context) error {
	var u struct {
		Name string `json:"key" binding:"required"`
		Url  string `json:"url" binding:"required"`
	}

	if err := c.ShouldBindJSON(&u); err != nil {
		return err
	}

	_, err := url.Parse(u.Url)
	if err != nil {
		return err
	}

	if _, ok := s.noticeUrl[u.Name]; !ok {
		s.noticeUrl[u.Name] = u.Url
	}

	return nil
}
