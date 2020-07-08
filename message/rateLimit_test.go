package message

import (
	"github.com/magiconair/properties/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"testing"
	"time"
)

func TestSec(t *testing.T) {
	rate := newRateLimit(r)

	assert.Equal(t, time.Second, rate.msgSec)
	assert.Equal(t, 10*time.Second, rate.sameSec)
}

func TestPerSec(t *testing.T) {
	rate := newRateLimit(r)

	testCase := []struct {
		mid []int64
		err []error
	}{
		{
			mid: []int64{1},
			err: []error{nil},
		},
		{
			mid: []int64{2, 2},
			err: []error{nil, errors.ErrRateMsg},
		},
	}

	for _, v := range testCase {
		for i, id := range v.mid {
			err := rate.perSec(id)
			assert.Equal(t, v.err[i], err)
		}
	}
}

func TestIsSameMsg(t *testing.T) {
	rate := newRateLimit(r)

	testCase := []struct {
		uid       []string
		IsSameMsg []error
	}{
		{
			uid:       []string{"1"},
			IsSameMsg: []error{nil},
		},
		{
			uid:       []string{"2", "2"},
			IsSameMsg: []error{nil, nil},
		},
		{
			uid:       []string{"3", "3", "3"},
			IsSameMsg: []error{nil, nil, errors.ErrRateSameMsg},
		},
		{
			uid:       []string{"4", "4", "4", "4"},
			IsSameMsg: []error{nil, nil, errors.ErrRateSameMsg, errors.ErrRateSameMsg},
		},
	}

	for _, v := range testCase {
		for i, u := range v.uid {
			err := rate.sameMsg("test", u)
			assert.Equal(t, v.IsSameMsg[i], err)
		}
	}
}
