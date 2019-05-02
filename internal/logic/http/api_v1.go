package http

import (
	stdErrors "errors"
	"fmt" /*should remove finally*/

	"github.com/gin-gonic/gin"
)

type apiCommon interface {
	mapParamters(c *gin.Context) error
	payLoadToJson() map[string]interface{}
	payLoadTransfer()
	businessLogic()
}

//RequestCommonPayLoad ...
//通用 request payload
type RequestCommonPayLoad struct {
	action    string
	from      string
	to        string
	query     string
	content   string
	Type      string
	Room      string
	Op, Speed int32
	Mids      []int64
	msg       []byte
}

//Jinyan ...
//禁言
type Jinyan struct {
	RequestCommonPayLoad
}

//MapParamters ...
//resolveURL implementation
func (j *Jinyan) mapParamters(c *gin.Context) error {

	if c.Param("from") == "" /* && TODO: search white list */ {
		return stdErrors.New("房間參數出錯")
	}
	if c.Param("to") == "" /* && TODO: search white list */ {
		return stdErrors.New("指定用戶出錯")
	}

	j.action = "禁言"
	j.from = c.Param("from")
	j.to = c.Param("to")
	j.content = c.Param("content")
	j.msg = []byte(c.Param("content"))

	return nil
}

//PayLoadToJSON ...
// 將request payload 轉成 json
func (j *Jinyan) payLoadToJSON() map[string]interface{} {
	//TODO: 需要實做 check error
	return gin.H{
		"action":  j.action,
		"from":    j.from,
		"to":      j.to,
		"content": j.content,
		"Type":    j.Type,
		"Room":    j.Room,
		"Op":      j.Op,
		"Speed":   j.Speed,
	}
}

// PayLoadTransfer ...
// 將 payload 轉成goim domain 資料
func (j *Jinyan) payLoadTransfer() {
	//TODO: 需要實做 check error
	j.Type = "Type 123"
	j.Room = "Room 234"
	j.Op = 123
	j.Speed = 123
	j.Mids = []int64{1, 2, 3, 4}
}

// BusinessLogic ...
// 禁言專用 BusinessLogic 處理
func (j *Jinyan) businessLogic() {
	//TODO: 需要實做 check error
	fmt.Println("禁言專用 BusinessLogic 處理 ... ")
}

func (s *Server) base(c *gin.Context) {
	//讀取 request uri binding

	//依 API route 規定驗證 uri binding

	//資料轉換

	//處理request相關的商業邏輯

	//將處理訊息丟向 logic

	//回覆 Response
	result(c, gin.H{"health": "I am fine"}, OK)
}

func (s *Server) jinyan(c *gin.Context) {

	//商業邏輯物件
	jinyan := new(Jinyan)

	//從 URI 取值	
	if err := jinyan.mapParamters(c); err != nil {
		errors(c, RequestErr, err.Error())
		return
	}

	//資料轉換
	jinyan.payLoadTransfer()

	//處理request相關的商業邏輯
	jinyan.businessLogic()

	//將處理訊息丟向 logic
	//TODO

	//回覆 Response
	result(c, jinyan.payLoadToJSON(), OK)
}
func (s *Server) fengsuo(c *gin.Context) {

}
func (s *Server) hongbao(c *gin.Context) {

}
func (s *Server) faker(c *gin.Context) {

}
