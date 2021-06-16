package time

import (
	"testing"
	"time"
)

func TestDurationText(t *testing.T) {
	var (
		input  = []byte("10s")
		output = time.Second * 10
		d      Duration
	)
	if err := d.UnmarshalText(input); err != nil {
		t.FailNow()
	}
	if int64(output) != int64(d) {
		t.FailNow()
	}
}

var g uint64

func BenchemarkDu(b *testing.B) {
	var l uint64
	for i := 0; i < b.N; i++ {
		//pprofTimer()
		g = l
	}
}
