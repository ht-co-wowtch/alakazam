package request

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"gitlab.com/jetfueltw/cpw/alakazam/pkg/bufio"
	pd "gitlab.com/jetfueltw/cpw/alakazam/protocol"
	"gitlab.com/jetfueltw/cpw/alakazam/protocol/grpc"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/business"
	"gitlab.com/jetfueltw/cpw/alakazam/test/protocol"
	"golang.org/x/net/websocket"
	"math/rand"
	"time"
)

type AuthToken struct {
	RoomID string `json:"room_id"`
	Token  string `json:"token"`
}

type ConnectReply struct {
	Uid        string              `json:"Uid"`
	Key        string              `json:"Key"`
	Permission business.Permission `json:"permission"`
}

type Auth struct {
	Uid   string
	Key   string
	Wr    *bufio.Writer
	Rd    *bufio.Reader
	Proto *grpc.Proto
	Reply *ConnectReply
}

func init() {
	rand.Seed(time.Now().Unix())
}

func Dial() (conn *websocket.Conn, err error) {
	conn, err = websocket.Dial("ws://127.0.0.1:3102/sub", "", "http://127.0.0.1")
	return
}

func DialAuth(roomId string) (auth Auth, err error) {
	return DialAuthToken(roomId, uuid.New().String())
}

func DialAuthToken(roomId, token string) (auth Auth, err error) {
	authToken := AuthToken{
		RoomID: roomId,
		Token:  token,
	}
	var (
		conn *websocket.Conn
	)

	conn, err = Dial()
	if err != nil {
		return
	}

	wr := bufio.NewWriter(conn)
	rd := bufio.NewReader(conn)

	proto := new(grpc.Proto)
	proto.Op = pd.OpAuth
	proto.Body, _ = json.Marshal(authToken)

	fmt.Printf("send Auth: %s\n", proto.Body)
	if err = protocol.Write(wr, proto); err != nil {
		return
	}
	if err = protocol.Read(rd, proto); err != nil {
		return
	}
	fmt.Printf("Auth Reply: %s\n", proto.Body)

	auth.Wr = wr
	auth.Rd = rd
	auth.Proto = proto

	reply := new(ConnectReply)

	if err = json.Unmarshal(proto.Body, &reply); err != nil {
		return
	}
	auth.Uid = string(reply.Uid)
	auth.Key = reply.Key
	auth.Reply = reply
	return
}
