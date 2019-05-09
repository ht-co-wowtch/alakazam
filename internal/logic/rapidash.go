package logic

import (
	"strconv"
	"time"
)

func renew(token string) (string, string) {
	u := time.Now().Unix()
	return strconv.Itoa(int(u)), "test"
}
