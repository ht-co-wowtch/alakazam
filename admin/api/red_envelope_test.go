package api

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http/httptest"
	"testing"
)

func TestGiveRedEnvelope(t *testing.T) {
	s := &Server{}
	g := gin.New()
	s.InitRoute(g)

	req := httptest.NewRequest("POST", "/red-envelope", nil)
	w := httptest.NewRecorder()
	g.ServeHTTP(w, req)
	result := w.Result()
	defer result.Body.Close()

	body, _ := ioutil.ReadAll(result.Body)
	expected := `{"count":10,"expire_at":"2019-07-02T17:40:50+08:00","id":"6aa077f3db794f14becc653aa788f3aa","message":"恭喜發財，大吉大利","publish_at":"2019-07-02T17:40:50+08:00","token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1NjIwNjc2NTAsImlkIjoiNmFhMDc3ZjNkYjc5NGYxNGJlY2M2NTNhYTc4OGYzYWEiLCJtZXNzYWdlIjoi5oGt5Zac55m86LKh77yM5aSn5ZCJ5aSn5YipIiwidWlkIjoiMTVjM2M2MWY5MDBhNDMzZmI4ZjFhOWIzMTE0Y2Y3MmMifQ.Q-NDcZ7qA0ftCamCDthp9kQR0rhjX2fZ3Ki-zLEfWR0","total_amount":10,"type":"equally","uid":"15c3c61f900a433fb8f1a9b3114cf72c"}`

	assert.Equal(t, expected, string(body))
}
