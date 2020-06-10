package money

import (
	"bytes"
	"fmt"
	"strings"
)

func FormatFloat64(v float64) string {
	return formatNumberString(fmt.Sprintf("%.2f", v))
}

func formatNumberString(v string) string {
	lastIndex := strings.Index(v, ".") - 1

	if lastIndex < 0 {
		lastIndex = len(v) - 1
	}

	var buffer []byte
	var strBuffer bytes.Buffer

	j := 0
	for i := lastIndex; i >= 0; i-- {
		j++
		buffer = append(buffer, v[i])

		if j == 3 && i > 0 && !(i == 1 && v[0] == '-') {
			buffer = append(buffer, ',')
			j = 0
		}
	}

	for i := len(buffer) - 1; i >= 0; i-- {
		strBuffer.WriteByte(buffer[i])
	}

	return strBuffer.String() + v[lastIndex+1:]
}
