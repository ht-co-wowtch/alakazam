package message

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

var (
	fakeMessageByte1 = `{"id":2802,"uid":"0d641b03d4d548dbb3a73a2197811261","type":"message","name":"","avatar":"other","message":"wdwdwdwd","time":"14:57:25","timestamp":1567580245}`
	fakeMessageByte2 = `{"id":2801,"uid":"0d641b03d4d548dbb3a73a2197811261","type":"message","name":"","avatar":"other","message":"wdwdwd","time":"14:56:37","timestamp":1567580197}`
	fakeMessageByte3 = `{"id":2601,"uid":"0d641b03d4d548dbb3a73a2197811261","type":"message","name":"","avatar":"other","message":"wdwd","time":"14:51:30","timestamp":1567579890}`
)

func TestHistoryGetV2(t *testing.T) {
	h, cache := mockHistory()

	cache.m.On("getMessage", mock.Anything, mock.Anything).Return([]string{
		fakeMessageByte1, fakeMessageByte2, fakeMessageByte3,
	}, nil)

	msg, err := h.GetV2(1, time.Now())

	assert.Nil(t, err)
	assert.Equal(t, []interface{}{stringJson(fakeMessageByte3), stringJson(fakeMessageByte2), stringJson(fakeMessageByte1)}, msg)
}

func TestMarshalJsonByString(t *testing.T) {
	str := []stringJson{
		stringJson(fakeMessageByte3),
		stringJson(fakeMessageByte2),
		stringJson(fakeMessageByte1),
	}

	b, err := json.Marshal(str)

	assert.Nil(t, err)
	assert.Equal(t, `[{"id":2601,"uid":"0d641b03d4d548dbb3a73a2197811261","type":"message","name":"","avatar":"other","message":"wdwd","time":"14:51:30","timestamp":1567579890},{"id":2801,"uid":"0d641b03d4d548dbb3a73a2197811261","type":"message","name":"","avatar":"other","message":"wdwdwd","time":"14:56:37","timestamp":1567580197},{"id":2802,"uid":"0d641b03d4d548dbb3a73a2197811261","type":"message","name":"","avatar":"other","message":"wdwdwdwd","time":"14:57:25","timestamp":1567580245}]`, string(b))
}

func mockHistory() (*History, *mockCache) {
	cache := new(mockCache)
	return &History{
		cache: cache,
	}, cache
}

type mockCache struct {
	m mock.Mock
}

func (m mockCache) getMessage(rid int32, at time.Time) ([]string, error) {
	arg := m.m.Called(rid, at)
	return arg.Get(0).([]string), arg.Error(1)
}
