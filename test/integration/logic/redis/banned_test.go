package redis

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestBanned(t *testing.T) {
	err := d.SetBanned("123", 3)
	assert.Nil(t, err)

	ex, ok, err := d.GetBanned("123")
	assert.Nil(t, err)
	assert.True(t, ok)
	assert.False(t, ex.IsZero())

	time.Sleep(time.Second * 3)

	ex, ok, err = d.GetBanned("123")
	assert.Nil(t, err)
	assert.False(t, ok)
	assert.True(t, ex.IsZero())
}

func TestDelBanned(t *testing.T) {
	err := d.SetBanned("456", 3)
	assert.Nil(t, err)

	err = d.DelBanned("456")
	assert.Nil(t, err)

	ex, ok, err := d.GetBanned("456")
	assert.Nil(t, err)
	assert.False(t, ok)
	assert.True(t, ex.IsZero())
}
