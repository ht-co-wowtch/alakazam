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
	msgA := Messages{
		Uid:     "1",
		Message: "test",
	}
	msgB := Messages{
		Uid:     "2",
		Message: "test",
	}
	msgC := Messages{
		Uid:     "3",
		Message: "test",
	}
	msgD := Messages{
		Uid:     "4",
		Message: "test",
	}

	testCase := []struct {
		msg       []Messages
		IsSameMsg []error
	}{
		{
			msg:       []Messages{msgA},
			IsSameMsg: []error{nil},
		},
		{
			msg:       []Messages{msgB, msgB},
			IsSameMsg: []error{nil, nil},
		},
		{
			msg:       []Messages{msgC, msgC, msgC},
			IsSameMsg: []error{nil, nil, errors.ErrRateSameMsg},
		},
		{
			msg:       []Messages{msgD, msgD, msgD, msgD},
			IsSameMsg: []error{nil, nil, errors.ErrRateSameMsg, errors.ErrRateSameMsg},
		},
	}

	for _, v := range testCase {
		for i, m := range v.msg {
			err := rate.sameMsg(m)
			assert.Equal(t, v.IsSameMsg[i], err)
		}
	}
}
