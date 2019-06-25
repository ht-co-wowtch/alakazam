package validation

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/go-playground/validator.v8"
	"testing"
)

var v *validator.Validate

func init() {
	v = validator.New(&validator.Config{TagName: "validate"})
}

func TestMin(t *testing.T) {
	var p = struct {
		C string `validate:"min=1"`
	}{}
	p.C = ""

	actual := validate(p)

	assert.Equal(t, "The field must be at least 1", actual["C"])
}

func TestMinPass(t *testing.T) {
	var p = struct {
		C string `validate:"min=1"`
	}{}
	p.C = "1"

	actual := v.Struct(p)

	assert.Nil(t, actual)
}

func TestMax(t *testing.T) {
	var p = struct {
		C string `validate:"max=1"`
	}{}
	p.C = "12"

	actual := validate(p)

	assert.Equal(t, "The field may not be greater than 1", actual["C"])
}

func TestMaxPass(t *testing.T) {
	var p = struct {
		C string `validate:"max=1"`
	}{}
	p.C = "1"

	actual := v.Struct(p)

	assert.Nil(t, actual)
}

func TestRequired(t *testing.T) {
	var p = struct {
		C int `validate:"required"`
	}{}

	actual := validate(p)

	assert.Equal(t, "The field field is required", actual["C"])
}

func TestRequiredPass(t *testing.T) {
	var p = struct {
		C int `validate:"required"`
	}{}

	p.C = 1

	actual := v.Struct(p)

	assert.Nil(t, actual)
}

func TestLen(t *testing.T) {
	var p = struct {
		C string `validate:"len=1"`
	}{}
	p.C = "12"

	actual := validate(p)

	assert.Equal(t, "This field must be 1", actual["C"])
}

func TestLenPass(t *testing.T) {
	var p = struct {
		C string `validate:"len=1"`
	}{}
	p.C = "1"

	actual := v.Struct(p)

	assert.Nil(t, actual)
}

func TestLt(t *testing.T) {
	var p = struct {
		C int `validate:"lt=10"`
	}{}
	p.C = 11

	actual := validate(p)

	assert.Equal(t, "The field must be less than 10", actual["C"])
}

func TestLtPass(t *testing.T) {
	var p = struct {
		C int `validate:"lt=10"`
	}{}
	p.C = 9

	actual := v.Struct(p)

	assert.Nil(t, actual)
}

func TestLte(t *testing.T) {
	var p = struct {
		C int `validate:"lte=10"`
	}{}
	p.C = 11

	actual := validate(p)

	assert.Equal(t, "The field must be less than or equal 10", actual["C"])
}

func TestLtePass(t *testing.T) {
	var p = struct {
		C int `validate:"lte=10"`
	}{}
	p.C = 10

	actual := v.Struct(p)

	assert.Nil(t, actual)
}

func TestGt(t *testing.T) {
	var p = struct {
		C int `validate:"gt=10"`
	}{}
	p.C = 9

	actual := validate(p)

	assert.Equal(t, "The field must be greater than 10", actual["C"])
}

func TestGtPass(t *testing.T) {
	var p = struct {
		C int `validate:"gt=10"`
	}{}
	p.C = 11

	actual := v.Struct(p)

	assert.Nil(t, actual)
}

func TestGte(t *testing.T) {
	var p = struct {
		C int `validate:"gte=10"`
	}{}
	p.C = 9

	actual := validate(p)

	assert.Equal(t, "The field must be greater than or equal 10", actual["C"])
}

func TestGtePass(t *testing.T) {
	var p = struct {
		C int `validate:"gte=10"`
	}{}
	p.C = 10

	actual := v.Struct(p)

	assert.Nil(t, actual)
}

func validate(data interface{}) map[string]string {
	err := v.Struct(data)
	e := err.(validator.ValidationErrors)
	return ValidationErrorsMap(e)
}
