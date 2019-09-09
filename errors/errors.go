package errors

import (
	"encoding/json"
	"gitlab.com/jetfueltw/cpw/micro/errdefs"
	"gitlab.com/jetfueltw/cpw/micro/validation"
	"gopkg.in/go-playground/validator.v8"
)

var (
	// 								例外	00**

	// 								通用	10**
	ErrNoRows  = errdefs.NotFound(1001, "没有资料", nil)
	ErrExist   = errdefs.Conflict(1002, "资料已存在", nil)
	ErrIllegal = errdefs.InvalidParameter(1003, "消息包含被禁止的内容", nil)

	// 								認證/會員	20**
	ErrNoMember        = errdefs.NotFound(2001, "没有会员资料", nil)
	ErrTokenUid        = errdefs.Unauthorized(2002, "帐号资料认证失败", nil)
	ErrValidationToken = errdefs.Unauthorized(2003, "用户认证失败", nil)
	ErrClaimsToken     = errdefs.Unauthorized(2004, "用户认证失败", nil)
	ErrValidToken      = errdefs.Unauthorized(2005, "用户认证失败", nil)
	ErrLogin           = errdefs.Unauthorized(2006, "请先登入会员", nil)
	ErrAuthorization   = errdefs.Unauthorized(2007, "Unauthorized", nil)
	ErrMemberNoMessage = errdefs.Unauthorized(2008, "您在永久禁言状态，无法发言", nil)
	ErrMemberBanned    = errdefs.Unauthorized(2009, "您在禁言状态，无法发言", nil)
	ErrBlockade        = errdefs.Unauthorized(2010, "您在封鎖状态，无法进入聊天室", nil)

	// 								金額	30**

	// 								紅包	40**
	ErrPublishAt = errdefs.InvalidParameter(4001, "预定发送时间不能大于现在", nil)

	// 								房間	50**
	ErrNoRoom      = errdefs.NotFound(5001, "没有房间资料", nil)
	ErrRoomClose   = errdefs.NotFound(5002, "目前房间已关闭", nil)
	ErrRateMsg     = errdefs.TooManyRequests(5003, "1秒内只能发一则消息", nil)
	ErrRateSameMsg = errdefs.TooManyRequests(5004, "10秒内相同讯息3次，自动禁言10分钟", nil)
	// 5005
	ErrRoomLimit     = "您无法发言，当前发言条件：前%d天充值不少于%d元；打码量不少于%d元"
	ErrRoomNoMessage = errdefs.Unauthorized(5006, "聊天室目前禁言状态，无法发言", nil)
)

const (
	NoLogin = 12014041
	// 沒有token
	noAuthorizationBearer = 15022001
	// 資料格式錯誤
	invalidParameter = 15021002
	// 餘額不足
	balanceCode = 12024020
	// 房間不存在
	roomNotFoundCode = 15025001
	// 找不到會員資料
	memberNotFound = 12024041
	// 隨機紅包金額不能小於包數
	redEnvelopeAmount = 15023001
	// 紅包不存在
	redEnvelopeNotFoundCode = 15024001
	// 紅包已關閉
	redEnvelopeIsClose = 15024002
	// 紅包發佈時間不能小於當下
	redEnvelopePublishTime = 15024006
	// 紅包已發佈過
	redEnvelopePublishExist = 15024007
	// 紅包未發佈但已過期
	redEnvelopePublishExpire = 15024008
)

func init() {
	if err := errdefs.SetCode(1002); err != nil {
		panic(err)
	}
	errdefs.SetOutput(output{})
	errdefs.SetJsonOut(output{})
	errdefs.SetValidationOut(output{})
	errdefs.SetErrorCode(1002, 1003, 0000)

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
