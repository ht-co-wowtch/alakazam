package http

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"net/http"
)

// 設定禁言
func (s *httpServer) setBanned(c *gin.Context) error {
	params := struct {
		Uid     string `json:"uid" binding:"required,len=32"`
		Expired int    `json:"expired" binding:"required"`
	}{
		Uid: c.Param("uid"),
	}
	if err := c.ShouldBindJSON(&params); err != nil {
		return err
	}

	m, err := s.member.Get(c.GetString("uid"))
	if err != nil {
		return err
	}

	switch m.Type {
	case models.STREAMER, models.MANAGE:
	default:
		return errors.ErrForbidden
	}

	if err := s.member.SetBanned(params.Uid, params.Expired, false); err != nil {
		return err
	}
	c.Status(http.StatusNoContent)
	return nil
}

// 解除禁言
func (s *httpServer) removeBanned(c *gin.Context) error {
	params := struct {
		Uid string `json:"uid" binding:"required,len=32"`
	}{
		Uid: c.Param("uid"),
	}
	if err := binding.Validator.ValidateStruct(&params); err != nil {
		return err
	}

	m, err := s.member.Get(c.GetString("uid"))
	if err != nil {
		return err
	}

	switch m.Type {
	case models.STREAMER, models.MANAGE:
	default:
		return errors.ErrForbidden
	}

	if err := s.member.RemoveBanned(params.Uid); err != nil {
		return err
	}
	c.Status(http.StatusNoContent)
	return nil
}
