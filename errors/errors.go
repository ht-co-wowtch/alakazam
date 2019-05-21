package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
)

var (
	ConnectError                   = eNew(http.StatusBadRequest, 10024000, "进入聊天室失败")
	FailureError                   = eNew(http.StatusBadRequest, 10024001, "操作失败")
	RoomError                      = eNew(http.StatusBadRequest, 10024001, "没有在此房间")
	NoRowsError                 = eNew(http.StatusNotFound, 10024040, "没有资料")
	BlockadeError, BlockadeMessage = eNewB(http.StatusUnauthorized, 10024011, "您在封鎖状态，无法进入聊天室")
	LoginError                     = eNew(http.StatusUnauthorized, 10024012, "请先登入会员")
	BannedError                    = eNew(http.StatusUnauthorized, 10024013, "您在禁言状态，无法发言")
	DataError                      = eNew(http.StatusUnprocessableEntity, 10024220, "资料验证错误")
	SetRoomError                   = eNew(http.StatusUnprocessableEntity, 10024221, "")
	TypeError                      = eNew(http.StatusInternalServerError, 10025000, "应用程序错误")
)

type Error struct {
	Status  int    `json:"-"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func eNew(status int, code int, message string) Error {
	return Error{
		Status:  status,
		Code:    code,
		Message: message,
	}
}

func eNewB(status int, code int, message string) (Error, []byte) {
	e := Error{
		Status:  status,
		Code:    code,
		Message: message,
	}
	b, err := json.Marshal(e)
	if err != nil {
		panic(err)
	}
	return e, b
}

func (e Error) Error() string {
	return fmt.Sprintf("%d: %v", e.Code, e.Message)
}

func (e Error) Mes(msg string) Error {
	e.Message = msg
	return e
}
