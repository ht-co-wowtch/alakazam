package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const cookieName = "chat"

var (
	rooms []room

	users map[string]user

	host string

	port string
)

type room struct {
	Id string `form:"roomId" binding:"required,numeric"`

	Title string `form:"title" binding:"required"`

	Type string `form:"type" binding:"required"`

	Tag string `form:"tag"`

	tag [] string

	Introduction string `form:"text" binding:"required"`
}

type user struct {
	id   string
	name string
}

func init() {
	rooms = append(rooms, room{
		Title:        "聊天測試區",
		Id:           "1000",
		Type:         "chat",
		Tag:          "1,1000",
		tag:          []string{"1", "1000"},
		Introduction: "聊天測試區",
	})

	users = make(map[string]user, 10)
}

func main() {
	flag.StringVar(&host, "h", "127.0.0.1", "chat host")
	flag.StringVar(&port, "p", "2222", "chat port")
	flag.Parse()

	g := gin.Default()
	g.LoadHTMLGlob("./templates/*")

	g.GET("login", loginForm)
	g.POST("login", login)

	user := g.Group("/", checkLogin)
	user.GET("/", indexForm)
	user.GET("/add", addForm)
	user.POST("/add", add)
	user.GET("/room/:id/:type/:tag", roomForm)
	user.POST("/push/:id/:type/:operation", push)
	user.POST("/pushAll", pushAll)
	user.GET("/push/:type", pushForm)

	g.Run(":2222")
}

func pushForm(c *gin.Context) {
	tag := []string{}
	id := []string{}
	t := []string{}

	for _, v := range rooms {
		tag = append(tag, v.tag...)
	}

	for _, v := range rooms {
		id = append(id, v.Id)
	}

	for _, v := range rooms {
		t = append(t, v.Type)
	}

	c.HTML(http.StatusOK, "push.html", gin.H{
		"push": c.Param("type"),
		"tag":  uniqueSlice(tag),
		"id":   id,
		"type": t,
		"host": host,
		"port": port,
	})
}

func push(c *gin.Context) {
	i, _ := c.Cookie(cookieName)
	if u, ok := users[i]; ok {
		text := fmt.Sprintf(`{"name":"%s", "content":"%s"}`, u.name, c.PostForm("text"))

		url := fmt.Sprintf("http://127.0.0.1:3111/goim/push/room?operation=%s&type=%s&room=%s",
			c.Param("operation"),
			c.Param("type"),
			c.Param("id"),
		)

		if _, err := http.DefaultClient.Post(url, "", strings.NewReader(text)); err == nil {
			c.JSON(http.StatusNoContent, gin.H{})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{})
		}
	}
}

func pushAll(c *gin.Context) {
	text := fmt.Sprintf(`{"name":"公告", "content":"%s"}`, c.PostForm("text"))
	url := []string{}
	key := c.PostForm("key")

	switch c.PostForm("push") {
	case "all":
		url = []string{"http://127.0.0.1:3111/goim/push/all?operation=1"}
	case "id":
		for _, v := range rooms {
			if v.Id == key {
				url = []string{fmt.Sprintf("http://127.0.0.1:3111/goim/push/room?operation=%s&type=%s&room=%s",
					v.Id,
					v.Type,
					v.Id,
				)}
			}
		}
	case "type":
		for _, v := range rooms {
			if v.Type == key {
				u := fmt.Sprintf("http://127.0.0.1:3111/goim/push/room?operation=%s&type=%s&room=%s",
					v.Id,
					v.Type,
					v.Id,
				)

				url = append(url, u)
			}
		}
	case "tag":
		url = []string{fmt.Sprintf("http://127.0.0.1:3111/goim/push/all?operation=%s", key)}
	}

	if len(url) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	for _, u := range url {
		fmt.Println(u)
		fmt.Println(text)
		if _, err := http.DefaultClient.Post(u, "", strings.NewReader(text)); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{})
			return
		}
	}

	c.JSON(http.StatusNoContent, gin.H{})
}

func roomForm(c *gin.Context) {
	i, _ := c.Cookie(cookieName)
	if u, ok := users[i]; ok {
		fmt.Println(c.Request.Host)
		c.HTML(http.StatusOK, "room.html", gin.H{
			"id":     c.Param("id"),
			"type":   c.Param("type"),
			"tag":    c.Param("tag"),
			"name":   u.name,
			"host":   host,
			"port":   port,
			"userId": u.id,
		})
	} else {
		c.Redirect(http.StatusMovedPermanently, "/login")
	}
}

func add(c *gin.Context) {
	r := room{}
	if err := c.Bind(&r); err == nil {
		for _, v := range rooms {
			if v.Id == r.Id {
				c.HTML(http.StatusOK, "add.html", gin.H{
					"msg": "房間號碼已存在",
				})
				return
			}
		}

		r.tag = []string{"1", r.Id}

		if r.Tag != "" {
			t := strings.Split(r.Tag, ",")

			for _, v := range t {
				if v == r.Id {
					c.HTML(http.StatusOK, "add.html", gin.H{
						"msg": "請不要輸入房間號碼",
					})

					return
				}
			}

			r.tag = append(r.tag, t...)
		}

		r.Tag = strings.Join(r.tag, ",")

		rooms = append(rooms, r)

		c.HTML(http.StatusOK, "add.html", gin.H{
			"msg": "新增成功",
		})

	} else {
		c.HTML(http.StatusOK, "add.html", gin.H{
			"msg": "新增失敗",
		})
	}
}

func addForm(c *gin.Context) {
	c.HTML(http.StatusOK, "add.html", gin.H{})
}

func login(c *gin.Context) {
	if name := c.DefaultPostForm("name", ""); name != "" {
		u := user{
			id:   strconv.FormatInt(time.Now().Unix(), 10),
			name: name,
		}
		users[u.id] = u

		c.SetCookie(cookieName, u.id, 3600, "/", c.Request.Host, false, true)
		c.Redirect(http.StatusMovedPermanently, "/")
	} else {
		c.Redirect(http.StatusMovedPermanently, "/login")
	}
}

func loginForm(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{})
}

func indexForm(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", rooms)
}

func checkLogin(c *gin.Context) {
	i, _ := c.Cookie(cookieName)
	if _, ok := users[i]; !ok {
		c.Header("Cache-Control", "no-store")
		c.Redirect(http.StatusMovedPermanently, "/login")
		c.Abort()
	}
}

func uniqueSlice(slice []string) []string {
	found := make(map[string]bool)
	total := 0
	for i, val := range slice {
		if _, ok := found[val]; !ok {
			found[val] = true
			(slice)[total] = (slice)[i]
			total++
		}
	}

	slice = (slice)[:total]

	return slice
}
