package validation

import (
	"bytes"
	"gopkg.in/go-playground/validator.v8"
)

func ValidationErrorsMap(err validator.ValidationErrors) map[string]string {
	buff := bytes.NewBufferString("")
	e := make(map[string]string, len(err))

	for _, v := range err {
		switch v.Tag {
		case "required":
			buff.WriteString("The field field is required")
		case "min":
			buff.WriteString("The field must be at least ")
			buff.WriteString(v.Param)
		case "max":
			buff.WriteString("The field may not be greater than ")
			buff.WriteString(v.Param)
		case "len":
			buff.WriteString("This field must be ")
			buff.WriteString(v.Param)
		case "lt":
			buff.WriteString("The field must be less than ")
			buff.WriteString(v.Param)
		case "lte":
			buff.WriteString("The field must be less than or equal ")
			buff.WriteString(v.Param)
		case "gt":
			buff.WriteString("The field must be greater than ")
			buff.WriteString(v.Param)
		case "gte":
			buff.WriteString("The field must be greater than or equal ")
			buff.WriteString(v.Param)
		}

		e[v.Name] = buff.String()

		buff.Reset()
	}

	return e
}
