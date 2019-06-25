package id

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUUid32(t *testing.T) {
	s := UUid32()

	assert.Len(t, s, 32)
}
