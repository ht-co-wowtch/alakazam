package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const cookieName = "chat"
const hosts = "http://127.0.0.1:3111"

var (
	rooms []room

	users map[string]user

	host string

	port string

	httpClient *http.Client
)

type room struct {
	Id string `form:"roomId" binding:"required,numeric"`

	Title string `form:"title" binding:"required"`

	Introduction string `form:"text" binding:"required"`
}

type user struct {
	id   string
	key  string
	name string
}

func init() {
	rooms = append(rooms, room{
		Title:        "聊天測試區",
		Id:           "1000",
		Introduction: "聊天測試區",
	})

	users = make(map[string]user, 10)

	httpClient = &http.Client{
		Timeout: time.Second * 5,
	}
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
	user.GET("/room/:id", roomForm)
	user.POST("/push/:id", push)
	user.POST("/pushAll", pushAll)
	user.GET("/push/:type", pushForm)
	user.GET("/count/:id", count)

	g.Run(":2222")
}

func pushForm(c *gin.Context) {
	id := []string{}

	for _, v := range rooms {
		id = append(id, v.Id)
	}

	c.HTML(http.StatusOK, "push.html", gin.H{
		"push": c.Param("type"),
		"id":   id,
		"host": host,
		"port": port,
	})
}

func push(c *gin.Context) {
	i, _ := c.Cookie(cookieName)
	if _, ok := users[i]; ok {
		_, err := pushRoom(c.PostForm("uid"), c.PostForm("key"), c.Param("id"), c.PostForm("text"))
		if err == nil {
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
		url = []string{"http://127.0.0.1:3111/push/all"}
	case "id":
		for _, v := range rooms {
			if v.Id == key {
				url = []string{fmt.Sprintf("http://127.0.0.1:3111/push/room?room=%s",
					v.Id,
				)}
			}
		}
	}

	if len(url) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	for _, u := range url {
		if _, err := http.DefaultClient.Post(u, "", strings.NewReader(text)); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{})
			return
		}
	}

	c.JSON(http.StatusNoContent, gin.H{})
}

func pushRoom(uid, key, roomId, message string) ([]byte, error) {
	data := url.Values{}
	data.Set("uid", uid)
	data.Set("key", key)
	data.Set("room_id", roomId)
	data.Set("message", message)
	return pushs(hosts+"/push/room", data)
}

func pushBroadcast(uid, key, message string) ([]byte, error) {
	data := url.Values{}
	data.Set("uid", uid)
	data.Set("key", key)
	data.Set("message", message)
	return pushs(fmt.Sprintf(hosts+"/push/all"), data)
}

func pushs(url string, data url.Values) (body []byte, err error) {
	resp, err := httpPost(url, data)
	if err != nil {
		return
	}

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = resp.Body.Close()
	fmt.Printf("response %s\n", string(body))
	return
}

func httpPost(url string, body url.Values) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, strings.NewReader(body.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func roomForm(c *gin.Context) {
	i, _ := c.Cookie(cookieName)
	if u, ok := users[i]; ok {
		c.HTML(http.StatusOK, "room.html", gin.H{
			"id":     c.Param("id"),
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

type co struct {
	Code int `json:"code"`
	Data []rco
}

type rco struct {
	Count int `json:"count"`

	RoomId string `json:"room_id"`
}

func count(c *gin.Context) {
	r, err := http.DefaultClient.Get(fmt.Sprintf("http://127.0.0.1:3111/online/top?limit=%d", c.Param("id")))

	if err == nil {
		defer r.Body.Close()

		b, _ := ioutil.ReadAll(r.Body)
		s := co{}
		json.Unmarshal(b, &s)

		if s.Code == 0 {
			for _, v := range s.Data {
				if v.RoomId == c.Param("id") {
					c.JSON(http.StatusOK, gin.H{
						"count": v.Count,
					})

					return
				}
			}
		}
	}
	c.JSON(http.StatusBadRequest, gin.H{})
}
