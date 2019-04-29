package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
)

var (
	userId   int
	roomId   int
	roomType string
	tag      string
)

func main() {
	flag.IntVar(&userId, "user", 0, "user id")
	flag.IntVar(&roomId, "room", 0, "room id")
	flag.StringVar(&roomType, "type", "", "room type")
	flag.StringVar(&tag, "tag", "", "room tag")
	flag.Parse()

	t := fmt.Sprintf(`{"mid":%d, "room_id":"%s://%d", "platform":"web", "accepts":[%s]}`,
		userId,
		roomType,
		roomId,
		tag,
	)
	token := []byte(t)

	c, _, err := websocket.DefaultDialer.Dial("ws://127.0.0.1:3102/sub", nil)
	if err != nil {
		panic(fmt.Sprintf("dial: %s", err))
	}

	buf := make([]byte, 16)

	binary.BigEndian.PutUint32(buf[0:], uint32(16+len(token)))
	binary.BigEndian.PutUint16(buf[4:], 16)
	binary.BigEndian.PutUint16(buf[6:], 1)
	binary.BigEndian.PutUint32(buf[8:], 7)
	binary.BigEndian.PutUint32(buf[12:], 1)
	b := bytes.NewBuffer(buf)
	b.Write(token)
	c.WriteMessage(websocket.BinaryMessage, b.Bytes())

	done := make(chan bool)
	defer func() {
		c.Close()
		close(done)
	}()

	go read(done, c)
	<-done
}

func read(done chan bool, c *websocket.Conn) {
	for {
		_, buf, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			done <- true
			return
		}

		op := binary.BigEndian.Uint32(buf[8:12])
		switch op {
		case 8:
			fmt.Println("進入房間成功")
		case 9:
			pl := binary.BigEndian.Uint32(buf[0:4])
			fmt.Println(string(buf[32:pl]))
		default:
		}
	}
}
