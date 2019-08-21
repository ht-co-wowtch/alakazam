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
		IsSameMsg []bool
	}{
		{
			msg:       []Messages{msgA},
			IsSameMsg: []bool{false},
		},
		{
			msg:       []Messages{msgB, msgB},
			IsSameMsg: []bool{false, false},
		},
		{
			msg:       []Messages{msgC, msgC, msgC},
			IsSameMsg: []bool{false, false, true},
		},
		{
			msg:       []Messages{msgD, msgD, msgD, msgD},
			IsSameMsg: []bool{false, false, true, true},
		},
	}

	for _, v := range testCase {
		for i, m := range v.msg {
			is, err := rate.IsSameMsg(m)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, v.IsSameMsg[i], is)
		}
	}

}
