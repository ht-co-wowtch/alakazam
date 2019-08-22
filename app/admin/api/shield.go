package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type shieldReq struct {
	Id      int    `json:"id"`
	Context string `json:"context" binding:"required"`
}

func (s *httpServer) CreateShield(c *gin.Context) error {
	var shirld shieldReq
	if err := c.ShouldBindJSON(&shirld); err != nil {
		return err
	}
	id, err := s.shield.Create(shirld.Context)
	if err != nil {
		return err
	}
	c.JSON(http.StatusOK, gin.H{
		"id": id,
	})
	return nil
}

func (s *httpServer) UpdateShield(c *gin.Context) error {
	var shirld shieldReq
	if err := c.ShouldBindJSON(&shirld); err != nil {
		return err
	}
	if err := s.shield.Update(shirld.Id, shirld.Context); err != nil {
		return err
	}
	c.Status(http.StatusNoContent)
	return nil
}

func (s *httpServer) DeleteShield(c *gin.Context) error {
	id := c.Param("id")
	i, err := strconv.Atoi(id)
	if err != nil {
		return err
	}
	if err := s.shield.Delete(i); err != nil {
		return err
	}
	c.Status(http.StatusNoContent)
	return nil
}
