package test

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	pd "gitlab.com/jetfueltw/cpw/alakazam/protocol"
	"gitlab.com/jetfueltw/cpw/alakazam/protocol/grpc"
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic"
	"gitlab.com/jetfueltw/cpw/alakazam/test/protocol"
	"gitlab.com/jetfueltw/cpw/alakazam/test/request"
	"testing"
	"time"
)

type resp struct {
	a        request.Auth
	p        []grpc.Proto
	response request.Response

	otherProto []grpc.Proto
	otherErr   error
}

// 讀取房間訊息
func TestReadRoomMessage(t *testing.T) {
	pushTest(t, "1000", "1000", func(a request.Auth) request.Response {
		return request.PushRoom(a.Uid, a.Key, "測試")
	}, func(r resp) {
		assert.Equal(t, pd.OpBatchRaw, r.a.Proto.Op)
		assert.Len(t, r.p, 1)
		assert.Nil(t, r.otherErr)
		assert.Len(t, r.otherProto, 1)
	})
}

// 讀取房間訊息格式
func TestReadRoomMessagePayload(t *testing.T) {
	pushTest(t, "2000", "2000", func(a request.Auth) request.Response {
		return request.PushRoom(a.Uid, a.Key, "測試")
	}, func(r resp) {
		l := new(logic.Message)
		json.Unmarshal(r.p[0].Body, l)
		tz, _ := time.Parse("15:04:05", l.Time)
		assert.Equal(t, "test", l.Name)
		assert.Equal(t, "", l.Avatar)
		assert.Equal(t, "測試", l.Message)
		assert.False(t, tz.IsZero())
	})
}

// 讀取廣播房間訊息
func TestReadBroadcastMessage(t *testing.T) {
	pushTest(t, "3000", "3001", func(a request.Auth) request.Response {
		return request.PushBroadcast([]string{"3000", "3001"}, "測試")
	}, func(r resp) {
		assert.Equal(t, pd.OpBatchRaw, r.a.Proto.Op)
		assert.Len(t, r.p, 1)
		assert.Nil(t, r.otherErr)
		assert.Len(t, r.otherProto, 1)
	})
}

func pushTest(t *testing.T, roomId string, otherRoomId string, f func(a request.Auth) (request.Response), ass func(resp)) {
	a, err := request.DialAuth(roomId)
	if err != nil {
		assert.Fail(t, err.Error())
		return
	}

	var (
		other      request.Auth
		otherErr   error
		otherProto []grpc.Proto
	)

	go func() {
		other, otherErr = request.DialAuth(otherRoomId)
		otherProto, otherErr = protocol.ReadMessage(other.Rd, other.Proto)
	}()

	r := f(a)
	time.Sleep(time.Second * 2)
	var p []grpc.Proto
	if p, err = protocol.ReadMessage(a.Rd, a.Proto); err != nil {
		assert.Fail(t, err.Error())
		return
	}
	rr := resp{
		response:   r,
		p:          p,
		a:          a,
		otherErr:   otherErr,
		otherProto: otherProto,
	}
	ass(rr)
}
