package errdefs

import (
	"encoding/json"
	"errors"
	"gitlab.com/jetfueltw/cpw/alakazam/pkg/validation"
	"gopkg.in/go-playground/validator.v8"
	"net/http"
)

var ErrInternalServer = New(http.StatusInternalServerError, 0, "Internal server error")

// json結構來表示此error代表的意義
type Error struct {
	Status  int         `json:"-"`
	Code    int         `json:"code"`
	Message interface{} `json:"message"`
	Err     error       `json:"-"`
}

// 專案代碼前4位數
var projectCode = 0

// 產生Error
func Err(err error) Error {
	var statusCode int
	var code int
	var index int
	var message interface{}

	switch e := err.(type) {
	case ErrInvalidParameter:
		statusCode = http.StatusBadRequest
	case ErrUnauthorized:
		statusCode = http.StatusUnauthorized
	case ErrPayment:
		statusCode = http.StatusPaymentRequired
	case ErrForbidden:
		statusCode = http.StatusForbidden
	case ErrNotFound:
		statusCode = http.StatusNotFound
	case ErrUnprocessableEntity:
		statusCode = http.StatusUnprocessableEntity
	case *Error:
		code = e.Code
		statusCode = http.StatusForbidden
	case ErrDataBase, ErrRedis:
		statusCode = http.StatusInternalServerError
	case validator.ValidationErrors:
		statusCode = http.StatusBadRequest
		err = UnprocessableEntity(err)
		message = validation.ValidationErrorsMap(e)
	case *json.UnmarshalTypeError:
		statusCode = http.StatusBadRequest
		message = "json: cannot unmarshal " + e.Value + " into field " + e.Field + " of type " + e.Type.String()
	default:
		statusCode = http.StatusInternalServerError
	}
	if e, ok := err.(causer); ok {
		index = e.Code()
		err = e.Cause()
	}
	if statusCode == http.StatusInternalServerError {
		message = "Internal server error"
	}
	if message == nil {
		message = err.Error()
	}
	if code == 0 {
		code = projectCode + (statusCode * 10) + index
	}
	return Error{
		Status:  statusCode,
		Code:    code,
		Message: message,
		Err:     err,
	}
}

var ErrCode = errors.New("Code must be thousands")

// error projectCode 前四位數
func SetCode(c int) error {
	if c < 999 || c > 9999 {
		return ErrCode
	}
	c *= 10000
	projectCode += c
	return nil
}

func New(status int, code int, message interface{}) Error {
	return Error{
		Status:  status,
		Code:    code,
		Message: message,
	}
}

func (e Error) Error() string {
	return e.Message.(string)
}

type Errors struct {
	Status  int               `json:"-"`
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Errors  map[string]string `json:"errors"`
}

func (e Errors) Error() string {
	return e.Message
}
