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
	msg := Display{
		Message: NullDisplayMessage{
			Text: "test",
		},
	}

	rate := newRateLimit(r)
	msgA := ProducerMessage{
		User: User{
			Uid: "1",
		},
		Display: msg,
	}
	msgB := ProducerMessage{
		User: User{
			Uid: "2",
		},
		Display: msg,
	}
	msgC := ProducerMessage{
		User: User{
			Uid: "3",
		},
		Display: msg,
	}
	msgD := ProducerMessage{
		User: User{
			Uid: "4",
		},
		Display: msg,
	}

	testCase := []struct {
		msg       []ProducerMessage
		IsSameMsg []error
	}{
		{
			msg:       []ProducerMessage{msgA},
			IsSameMsg: []error{nil},
		},
		{
			msg:       []ProducerMessage{msgB, msgB},
			IsSameMsg: []error{nil, nil},
		},
		{
			msg:       []ProducerMessage{msgC, msgC, msgC},
			IsSameMsg: []error{nil, nil, errors.ErrRateSameMsg},
		},
		{
			msg:       []ProducerMessage{msgD, msgD, msgD, msgD},
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
