package errdefs

import (
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"gopkg.in/go-playground/validator.v8"
	"net/http"
	"strconv"
	"testing"
)

var errTest = errors.New("test")

func TestSetCode(t *testing.T) {
	testCase := []struct {
		code       int
		assertCode int
		error      error
	}{
		{
			code:       1502,
			assertCode: 15020000,
		},
		{
			code:       1102,
			assertCode: 11020000,
		},
		{
			code:       1,
			assertCode: 0,
			error:      ErrCode,
		},
	}

	for i, v := range testCase {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			projectCode = 0
			err := SetCode(v.code)

			assert.Equal(t, v.error, err)
			assert.Equal(t, v.assertCode, projectCode)
		})
	}
}

func TestErr(t *testing.T) {
	v := validator.New(&validator.Config{TagName: "validate"})
	var p = struct {
		C string `validate:"min=1"`
	}{}
	p.C = ""

	errValidation := v.Struct(p)

	testCases := []struct {
		err        error
		statusCode int
		code       int
		message    interface{}
		Err        error
	}{
		{
			err:        errTest,
			statusCode: http.StatusInternalServerError,
			code:       http.StatusInternalServerError * 10,
			message:    "Internal server error",
			Err:        errTest,
		},
		{
			err:        InvalidParameter(errTest, 1),
			statusCode: http.StatusBadRequest,
			code:       http.StatusBadRequest*10 + 1,
			message:    "test",
			Err:        errTest,
		},
		{
			err:        Unauthorized(errTest, 2),
			statusCode: http.StatusUnauthorized,
			code:       http.StatusUnauthorized*10 + 2,
			message:    "test",
			Err:        errTest,
		},
		{
			err:        Payment(errTest, 3),
			statusCode: http.StatusPaymentRequired,
			code:       http.StatusPaymentRequired*10 + 3,
			message:    "test",
			Err:        errTest,
		},
		{
			err:        Forbidden(errTest, 4),
			statusCode: http.StatusForbidden,
			code:       http.StatusForbidden*10 + 4,
			message:    "test",
			Err:        errTest,
		},
		{
			err:        NotFound(errTest, 5),
			statusCode: http.StatusNotFound,
			code:       http.StatusNotFound*10 + 5,
			message:    "test",
			Err:        errTest,
		},
		{
			err:        UnprocessableEntity(errTest, 6),
			statusCode: http.StatusUnprocessableEntity,
			code:       http.StatusUnprocessableEntity*10 + 6,
			message:    "test",
			Err:        errTest,
		},
		{
			err:        DataBase(errTest, 8),
			statusCode: http.StatusInternalServerError,
			code:       http.StatusInternalServerError*10 + 8,
			message:    "Internal server error",
			Err:        errTest,
		},
		{
			err:        Redis(errTest, 9),
			statusCode: http.StatusInternalServerError,
			code:       http.StatusInternalServerError*10 + 9,
			message:    "Internal server error",
			Err:        errTest,
		},
		{
			err:        errValidation,
			statusCode: http.StatusBadRequest,
			code:       http.StatusBadRequest * 10,
			message:    map[string]string{"C": "The field must be at least 1"},
			Err:        errValidation,
		},
	}

	for _, v := range testCases {
		t.Run(http.StatusText(v.statusCode), func(t *testing.T) {
			e := Err(v.err)

			assert.Equal(t, Error{
				Status:  v.statusCode,
				Code:    v.code,
				Message: v.message,
				Err:     v.Err,
			}, e)
		})
	}
}

func TestJson(t *testing.T) {
	e := Error{}

	b, _ := json.Marshal(e)

	assert.Equal(t, `{"code":0,"message":null}`, string(b))

	e = Error{
		Message: map[string]string{
			"error": "test",
		},
	}

	b, _ = json.Marshal(e)

	assert.Equal(t, `{"code":0,"message":{"error":"test"}}`, string(b))
}

func TestNew(t *testing.T) {
	e := New(1, 0, "test")
	es := New(1, 0, map[string]string{"k": "v"})

	assert.Equal(t, Error{Status: 1, Code: 0, Message: "test"}, e)
	assert.Equal(t, Error{Status: 1, Code: 0, Message: map[string]string{"k": "v"}}, es)
}
