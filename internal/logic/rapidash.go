package logic

import "time"

func renew(token string) (int64, string) {
	return time.Now().Unix(), "test"
}
