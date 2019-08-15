package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (s *httpServer) giveRedEnvelope(c *gin.Context) error {
	c.JSON(http.StatusOK, gin.H{
		"id":           "6aa077f3db794f14becc653aa788f3aa",
		"uid":          "15c3c61f900a433fb8f1a9b3114cf72c",
		"type":         "equally",
		"message":      "恭喜發財，大吉大利",
		"total_amount": 10,
		"count":        10,
		"publish_at":   "2019-07-02T17:40:50+08:00",
		"expire_at":    "2019-07-02T17:40:50+08:00",
		"token":        "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1NjIwNjc2NTAsImlkIjoiNmFhMDc3ZjNkYjc5NGYxNGJlY2M2NTNhYTc4OGYzYWEiLCJtZXNzYWdlIjoi5oGt5Zac55m86LKh77yM5aSn5ZCJ5aSn5YipIiwidWlkIjoiMTVjM2M2MWY5MDBhNDMzZmI4ZjFhOWIzMTE0Y2Y3MmMifQ.Q-NDcZ7qA0ftCamCDthp9kQR0rhjX2fZ3Ki-zLEfWR0",
	})
	return nil
}
