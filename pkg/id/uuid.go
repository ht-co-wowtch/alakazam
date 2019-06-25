package id

import (
	"encoding/hex"
	"github.com/google/uuid"
)

func UUid32() string {
	buf := make([]byte, 32)
	b, _ := uuid.New().MarshalBinary()

	hex.Encode(buf, b)

	return string(buf)
}
