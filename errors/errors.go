package errors

import (
	"encoding/json"
	"gitlab.com/jetfueltw/cpw/micro/errdefs"
	"gitlab.com/jetfueltw/cpw/micro/validation"
	"gopkg.in/go-playground/validator.v8"
	"net/http"
)

var (
	ErrPublishAt = errdefs.InvalidParameter(New("预定发送时间不能大于现在"), 0)
	ErrExist     = errdefs.InvalidParameter(New("资料已存在"), 1)

	ErrLogin         = errdefs.Unauthorized(New("请先登入会员"))
	ErrReLogin       = errdefs.Unauthorized(New("请重新登入会员"))
	ErrRoomBanned    = errdefs.Unauthorized(New("聊天室目前禁言状态，无法发言"), 1)
	ErrBanned        = errdefs.Unauthorized(New("您在禁言状态，无法发言"), 2)
	ErrAuthorization = errdefs.Unauthorized(New("Unauthorized"), 3)

	ErrNoPage = errdefs.NotFound(New("无此Api"))
	ErrNoRows = errdefs.NotFound(New("没有资料"), 1)

	ErrRateMsg     = errdefs.TooManyRequests(New("1秒内只能发一则消息"), 1)
	ErrRateSameMsg = errdefs.TooManyRequests(New("10秒内相同讯息3次，自动禁言10分钟"), 2)

	ErrTokenUid = errdefs.Forbidden(New("帐号资料认证失败"), 1)

	//ConnectError = eNew(http.StatusBadRequest, 10024000, "进入聊天室失败")
	//FailureError = eNew(http.StatusBadRequest, 10024001, "操作失败")
	//
	//UserError = eNew(http.StatusBadRequest, 10024003, "取得用户资料失败")
	//
	//MoneyError   = eNew(http.StatusUnauthorized, 10024015, "您无法发言，当前发言条件：前%d天充值不少于%d元；打码量不少于%d元")
	//BalanceError = eNew(http.StatusPaymentRequired, 10024020, "您的余额不足发红包")
	//AmountError  = eNew(http.StatusPaymentRequired, 10024021, "金额错误")
	//DataError    = eNew(http.StatusUnprocessableEntity, 10024220, "资料验证错误")
)

const (
	// 紅包已過期
	TakeEnvelopeExpiredCode = 15024011

	// 紅包不存在
	EnvelopeNotFoundCode = 15024041

	// 餘額不足
	BalanceCode = 12024020

	// 房間不存在
	RoomNotFoundCode = 15024042

	// 預定發送時間過期
	PublishAtCode = 15024002
)

func init() {
	if err := errdefs.SetCode(1002); err != nil {
		panic(err)
	}
	errdefs.SetOutput(output{})

	validation.Set(validation.Required, "栏位必填")
	validation.Set(validation.Min, "栏位最大值或长度至少")
	validation.Set(validation.Max, "栏位最大值或长度最大")
	validation.Set(validation.Len, "栏位长度必须是")
	validation.Set(validation.Lt, "栏位必须小于")
	validation.Set(validation.Lte, "栏位必须小于或等于")
	validation.Set(validation.Gt, "栏位必须大于")
	validation.Set(validation.Gte, "栏位必须大于或等于")
}

type output struct{}

func (m output) Validation(e validator.ValidationErrors) interface{} {
	return validation.ValidationErrorsMap(e)
}

func (m output) JsonUnmarshalType(e *json.UnmarshalTypeError) interface{} {
	return map[string]string{
		e.Field: "栏位资料格式有误",
	}
}

func (m output) GetInternalServer() string {
	return "应用程序错误"
}

func (m output) Error(e *errdefs.Error) string {
	switch e.Code {
	case BalanceCode:
		return "您的余额不足发红包"
	case EnvelopeNotFoundCode:
		e.Status = http.StatusNotFound
		return "红包不存在"
	case RoomNotFoundCode:
		e.Status = http.StatusNotFound
		return "房间不存在"
	case PublishAtCode:
		e.Status = http.StatusNotFound
		return "预定发送时间不能大于现在"
	}
	return "操作失败"
}

func (m output) Other(err error) string {
	return ""
}

type Error struct {
	Message string
}

func New(message string) Error {
	return Error{Message: message}
}

func (e Error) Error() string {
	return e.Message
}
