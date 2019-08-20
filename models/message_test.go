package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestName(t *testing.T) {
	assert.Nil(t, prepareTestDatabase())

	data, err := s.GetRoomMessage(1)

	assert.Nil(t, err)
	assert.Nil(t, data)
}
