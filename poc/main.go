package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	"net/http"
)

var (
	host string
)

func main() {
	flag.StringVar(&host, "h", "127.0.0.1", "chat host")
	flag.Parse()

	g := gin.Default()
	g.LoadHTMLGlob("./templates/*")

	g.GET("/:id", roomForm)

	g.Run(":2222")
}

func roomForm(c *gin.Context) {
	c.HTML(http.StatusOK, "room.html", gin.H{
		"id":   c.Param("id"),
		"host": host,
	})
}
