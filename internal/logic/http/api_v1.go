package http

import (
	"context"
	stdErrors "errors"
	"fmt"
	"io/ioutil"

	"github.com/gin-gonic/gin"
)

type RequestCommonPayLoad struct {
	action, from, to, query, content, Type, Room string
	Op, Speed                                    int32
	Mids                                         []int64
	msg                                          []byte
}

// 以user key來推送訊息
func (s *Server) copyPushKeys(c *gin.Context) {
	var arg struct {
		Op   int32    `form:"operation"`
		Keys []string `form:"keys"`
	}
	if err := c.BindQuery(&arg); err != nil {
		errors(c, RequestErr, err.Error())
		return
	}
	msg, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		errors(c, RequestErr, err.Error())
		return
	}
	if err = s.logic.PushKeys(context.TODO(), arg.Op, arg.Keys, msg); err != nil {
		result(c, nil, RequestErr)
		return
	}
	result(c, nil, OK)
}

// 以user id來推送訊息
func (s *Server) copyPushMids(c *gin.Context) {
	var arg struct {
		Op   int32   `form:"operation"`
		Mids []int64 `form:"mids"`
	}
	if err := c.BindQuery(&arg); err != nil {
		errors(c, RequestErr, err.Error())
		return
	}
	msg, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		errors(c, RequestErr, err.Error())
		return
	}
	if err = s.logic.PushMids(context.TODO(), arg.Op, arg.Mids, msg); err != nil {
		errors(c, ServerErr, err.Error())
		return
	}
	result(c, nil, OK)
}

// 以operation來推送訊息
func (s *Server) copyPushAll(c *gin.Context) {
	var arg struct {
		Op    int32 `form:"operation" binding:"required"`
		Speed int32 `form:"speed"`
	}
	if err := c.BindQuery(&arg); err != nil {
		errors(c, RequestErr, err.Error())
		return
	}
	msg, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		errors(c, RequestErr, err.Error())
		return
	}
	if err = s.logic.PushAll(c, arg.Op, arg.Speed, msg); err != nil {
		errors(c, ServerErr, err.Error())
		return
	}
	result(c, nil, OK)
}

//禁言  以user id來推送訊息
//封鎖  以user id來推送訊息
//紅包 以operation來推送訊息
//假人 以room id來推送訊息

func (s *Server) base(c *gin.Context) {
	//讀取 request uri binding

	//依 API route 規定驗證 uri binding

	//資料轉換

	//處理request相關的商業邏輯

	//將處理訊息丟向 logic

	//回覆 Response
	result(c, gin.H{"health": "I am fine"}, OK)
}

/*

  //禁言
  v1.POST("/jinyan/:room/:uid/:content",s.jinyan)
  //封鎖
  v1.POST("/fengsuo/:room/:uid/:content",s.fengsuo)
  //紅包
  v1.POST("/hongbao/:room/:uid/:content",s.hongbao)
  //假人
  v1.POST("/faker/:room/:uid/:content",s.faker)


*/

// 以room id來推送訊息
func (s *Server) copyPushRoom(c *gin.Context) {
	var arg struct {
		Op   int32  `form:"operation" binding:"required"`
		Type string `form:"type" binding:"required"`
		Room string `form:"room" binding:"required"`
	}
	if err := c.BindQuery(&arg); err != nil {
		errors(c, RequestErr, err.Error())
		return
	}
	msg, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		errors(c, RequestErr, err.Error())
		return
	}
	if err = s.logic.PushRoom(c, arg.Op, arg.Type, arg.Room, msg); err != nil {
		errors(c, ServerErr, err.Error())
		return
	}
	result(c, nil, OK)
}

func fromURLMap(action string, c *gin.Context, payload *RequestCommonPayLoad) error {

	if action == "" /* && TODO: search white list */ {
		return stdErrors.New("action參數出錯")
	}
	if c.Param("from") == "" /* && TODO: search white list */ {
		return stdErrors.New("房間參數出錯")
	}
	if c.Param("to") == "" /* && TODO: search white list */ {
		return stdErrors.New("指定用戶出錯")
	}

	payload.action = action
	payload.from = c.Param("from")
	payload.to = c.Param("to")
	payload.content = c.Param("content")
	payload.msg = []byte(c.Param("content"))

	return nil
}

func payloadTransfer(load *RequestCommonPayLoad) {
	load.Type = "Type 123"
	load.Room = "Room 234"
	load.Op = 123
	load.Speed = 123
	load.Mids = []int64{1, 2, 3, 4}
}

func businessLogic(load *RequestCommonPayLoad) {
	fmt.Println("[business logic process]")
}

func dumpRequestPayLoad(p RequestCommonPayLoad) map[string]interface{} {
	return gin.H{
		"action":  p.action,
		"from":    p.from,
		"to":      p.to,
		"content": p.content,
		"Type":    p.Type,
		"Room":    p.Room,
		"Op":      p.Op,
		"Speed":   p.Speed,
	}
}

func (s *Server) jinyan(c *gin.Context) {

	var load RequestCommonPayLoad
	//讀取 request uri binding
	//依 API route 規定驗證 uri binding
	//TODO: check param goes here
	if err := fromURLMap("禁言", c, &load); err != nil {
		errors(c, RequestErr, err.Error())
		return
	}

	//資料轉換
	//TODO
	payloadTransfer(&load)

	//處理request相關的商業邏輯
	//TODO
	businessLogic(&load)

	//將處理訊息丟向 logic
	/*
		if err := s.logic.PushRoom(c, load.Op, load.Type, load.Room, load.msg); err != nil {
			errors(c, ServerErr, err.Error())
			return
		}
	*/
	//回覆 Response
	result(c, dumpRequestPayLoad(load), OK)
}
func (s *Server) fengsuo(c *gin.Context) {

}
func (s *Server) hongbao(c *gin.Context) {

}
func (s *Server) faker(c *gin.Context) {

}
