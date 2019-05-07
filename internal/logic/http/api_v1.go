package http

import (
	stdErrors "errors"
	"fmt" /*should remove finally*/

	"github.com/gin-gonic/gin"
)

//RequestPayLoad ...
//通用 request payload
type RequestPayLoad struct {
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
	RequestPayLoad
}

//MapParamters ...
//resolveURL implementation
func (me *RequestPayLoad) mapParamters(c *gin.Context) error {

	if c.Param("from") == "" /* && TODO: search white list */ {
		return stdErrors.New("房間參數出錯")
	}
	if c.Param("to") == "" /* && TODO: search white list */ {
		return stdErrors.New("指定用戶出錯")
	}

	me.action = "禁言"
	me.from = c.Param("from")
	me.to = c.Param("to")
	me.content = c.Param("content")
	me.msg = []byte(c.Param("content"))

	return nil
}

//PayLoadToJSON ...
// 將request payload 轉成 json
func (me *RequestPayLoad) payLoadToJSON() map[string]interface{} {
	//TODO: 需要實做 check error
	return gin.H{
		"action":  me.action,
		"from":    me.from,
		"to":      me.to,
		"content": me.content,
		"Type":    me.Type,
		"Room":    me.Room,
		"Op":      me.Op,
		"Speed":   me.Speed,
	}
}

// PayLoadTransfer ...
// 將 payload 轉成goim domain 資料
func (me *RequestPayLoad) payLoadTransfer() {
	//TODO: 需要實做 check error
	me.Type = "Type 123"
	me.Room = "Room 234"
	me.Op = 123
	me.Speed = 123
	me.Mids = []int64{1, 2, 3, 4}
}



// BusinessLogic ...
// 禁言專用 BusinessLogic 處理
func (j *Jinyan) businessLogic() {
	//TODO: 需要實做 check error
	fmt.Println("xxx禁言專用xxx BusinessLogic 處理 ... ")
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
