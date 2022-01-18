package errors

import (
	"encoding/json"

	"gitlab.com/ht-co/micro/errdefs"
	"gitlab.com/ht-co/micro/validation"
	"gopkg.in/go-playground/validator.v8"
)

const (
	NoLoginMessage = "请先登入会员"
	RoomBanned     = "聊天室目前禁言状态，无法发言"
	MemberBanned   = "您在永久禁言状态，无法发言"
)

var (
	// 								例外	00**

	// 								通用	10**
	ErrNoRows  = errdefs.NotFound(1001, "没有资料")
	ErrExist   = errdefs.Conflict(1002, "资料已存在")
	ErrIllegal = errdefs.InvalidParameter(1003, "消息包含被禁止的内容")

	// 								認證/會員	20**
	ErrNoMember        = errdefs.NotFound(2001, "没有会员资料")
	ErrTokenUid        = errdefs.Unauthorized(2002, "帐号资料认证失败")
	ErrLogin           = errdefs.Unauthorized(2006, NoLoginMessage)
	ErrAuthorization   = errdefs.Unauthorized(2007, "Unauthorized")
	ErrMemberNoMessage = errdefs.Unauthorized(2008, MemberBanned)
	ErrMemberBanned    = errdefs.Unauthorized(2009, "您在禁言状态，无法发言")
	ErrValidationToken = errdefs.Unauthorized(2003, "用户认证失败")
	ErrClaimsToken     = errdefs.Unauthorized(2004, "用户认证失败")
	ErrValidToken      = errdefs.Unauthorized(2005, "用户认证失败")
	ErrBlockade        = errdefs.Unauthorized(2010, "您在封鎖状态，无法进入聊天室")
	ErrForbidden       = errdefs.InvalidParameter(2011, "无操作权限")
	ErrNoOnline        = errdefs.Unauthorized(2012, "用户不在线上")

	// 								金額	30**

	// 								紅包	40**
	ErrPublishAt = errdefs.InvalidParameter(4001, "预定发送时间不能大于现在")

	// 								房間	50**
	ErrRateMsg     = errdefs.TooManyRequests(5003, "1秒内只能发一则消息")
	ErrRateSameMsg = errdefs.TooManyRequests(5004, "10秒内相同讯息3次，自动禁言10分钟")
	ErrNoRoom      = errdefs.NotFound(5001, "没有房间资料")
	ErrRoomClose   = errdefs.NotFound(5002, "目前房间已关闭")
	// 5005
	ErrRoomLimit     = "您无法发言，当前发言条件：前%d天充值不少于%d元；打码量不少于%d元"
	ErrRoomNoMessage = errdefs.Unauthorized(5006, RoomBanned)
)

const (
	NoLogin = 12014041
)

var errMessage = map[int]string{
	15022001: "无法认证身份",
	15021002: "资料格式错误",
	15025001: "房间不存在",
	12024020: "余额不足",
	12024041: "找不到会员资料",
	15023001: "红包金额不能小于包数",
	15024001: "红包不存在",
	15024002: "红包不存在",
	15024006: "红包发布时间不能小于当下",
	15024007: "红包已发布过",
	15024008: "红包未发布但已过期",
	15023004: "额度最多￥100000",
}

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

func (m output) Error(e *errdefs.Causer) string {
	if err, ok := errMessage[e.Code]; ok {
		return err
	}
	return "操作失败"
}

func (m output) GetInternalServer() string {
	return "应用程序错误"
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
