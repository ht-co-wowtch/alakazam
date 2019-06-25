package cache

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAddServerOnline(t *testing.T) {
	unix := time.Now().Unix()
	server := &Online{
		Server:    "123",
		RoomCount: map[string]int32{"1": 1, "2": 2},
		Updated:   unix,
	}
	err := c.AddServerOnline("123", server)

	assert.Nil(t, err)

	o, err := c.ServerOnline("123")

	assert.Equal(t, server, o)
}