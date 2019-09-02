package errors

import (
	"encoding/json"
	"gitlab.com/jetfueltw/cpw/micro/errdefs"
	"gitlab.com/jetfueltw/cpw/micro/validation"
	"gopkg.in/go-playground/validator.v8"
)

var (
	// 沒有資料
	ErrNoMember  = errdefs.NotFound(New("没有会员资料"), 4041)
	ErrNoRoom    = errdefs.NotFound(New("没有房间资料"), 4042)
	ErrRoomClose = errdefs.NotFound(New("目前房间已关闭"), 4043)

	// 限速
	ErrRateMsg     = errdefs.TooManyRequests(New("1秒内只能发一则消息"), 4291)
	ErrRateSameMsg = errdefs.TooManyRequests(New("10秒内相同讯息3次，自动禁言10分钟"), 4292)

	// 身份認證
	ErrValidationToken = errdefs.Unauthorized(New("用户认证失败"), 4011)
	ErrClaimsToken     = errdefs.Unauthorized(New("用户认证失败"), 4012)
	ErrValidToken      = errdefs.Unauthorized(New("用户认证失败"), 4013)
	ErrLogin           = errdefs.Unauthorized(New("请先登入会员"), 4014)
	// 4035
	ErrRoomLimit       = "您无法发言，当前发言条件：前%d天充值不少于%d元；打码量不少于%d元"
	ErrMemberNoMessage = errdefs.Unauthorized(New("您在永久禁言状态，无法发言"), 4015)
	ErrMemberBanned    = errdefs.Unauthorized(New("您在禁言状态，无法发言"), 4016)
	ErrRoomNoMessage   = errdefs.Unauthorized(New("聊天室目前禁言状态，无法发言"), 4017)
	ErrBlockade        = errdefs.Unauthorized(New("您在封鎖状态，无法进入聊天室"), 4018)

	// 系統異常
	ErrInternalServer = errdefs.InternalServer(New("操作失败，系统异常"), 5000)

	ErrPublishAt = errdefs.InvalidParameter(New("预定发送时间不能大于现在"), 0)
	ErrExist     = errdefs.InvalidParameter(New("资料已存在"), 1)

	ErrAuthorization = errdefs.Unauthorized(New("Unauthorized"), 3)

	ErrNoRows = errdefs.NotFound(New("没有资料"), 1)

	ErrTokenUid = errdefs.Unauthorized(New("帐号资料认证失败"), 1)
)

const (
	NoLogin = 12014041
	// 沒有token
	noAuthorizationBearer = 15024010
	// 資料格式錯誤
	invalidParameter = 15024220
	// 餘額不足
	balanceCode = 12024020
	// 房間不存在
	roomNotFoundCode = 15024042
	// 找不到會員資料
	memberNotFound = 12024041
	// 隨機紅包金額不能小於包數
	redEnvelopeAmount = 15021001
	// 紅包不存在
	redEnvelopeNotFoundCode = 15024044
	// 紅包已關閉
	redEnvelopeIsClose = 15024045
	// 紅包發佈時間不能小於當下
	redEnvelopePublishTime = 15024047
	// 紅包已發佈過
	redEnvelopePublishExist = 15024091
	// 紅包未發佈但已過期
	redEnvelopePublishExpire = 15024048
)

func init() {
	if err := errdefs.SetCode(1002); err != nil {
		panic(err)
	}
	errdefs.SetOutput(output{})
	errdefs.SetJsonOut(output{})
	errdefs.SetValidationOut(output{})

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

func (m output) GetValidationMessage() string {
	return "栏位资料格式有误"
}

func (m output) JsonUnmarshalType(e *json.UnmarshalTypeError) interface{} {
	return map[string]string{
		e.Field: "栏位资料格式有误",
	}
}

func (m output) GetJsonUnmarshalTypeMessage() string {
	return "栏位资料型态有误"
}

func (m output) GetInternalServer() string {
	return "应用程序错误"
}

func (m output) Error(e *errdefs.Causer) string {
	switch e.Code {
	case noAuthorizationBearer:
		return "无法认证身份"
	case invalidParameter:
		return "资料格式错误"
	case roomNotFoundCode:
		return "房间不存在"
	case balanceCode:
		return "余额不足"
	case memberNotFound:
		return "找不到会员资料"
	case redEnvelopeAmount:
		return "红包金额不能小于包数"
	case redEnvelopeNotFoundCode, redEnvelopeIsClose:
		return "红包不存在"
	case redEnvelopePublishTime:
		return "红包发布时间不能小于当下"
	case redEnvelopePublishExist:
		return "红包已发布过"
	case redEnvelopePublishExpire:
		return "红包未发布但已过期"
	}
	return "操作失败"
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
