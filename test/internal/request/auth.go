package request

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/permission"
	"gitlab.com/jetfueltw/cpw/alakazam/pkg/bufio"
	pd "gitlab.com/jetfueltw/cpw/alakazam/protocol"
	"gitlab.com/jetfueltw/cpw/alakazam/protocol/grpc"
	"gitlab.com/jetfueltw/cpw/alakazam/test/internal/protocol"
	"golang.org/x/net/websocket"
	"math/rand"
	"time"
)

type AuthToken struct {
	Uid    string `json:"uid"`
	Token  string `json:"token"`
	RoomID string `json:"room_id"`
}

type ConnectReply struct {
	Uid        string                `json:"Uid"`
	Key        string                `json:"Key"`
	Permission permission.Permission `json:"permission"`
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
	return DialAuthToken("82ea16cd2d6a49d887440066ef739669", roomId, uuid.New().String())
}

func DialAuthUser(uid, roomId string) (auth Auth, err error) {
	return DialAuthToken(uid, roomId, uuid.New().String())
}

func DialAuthToken(uid, roomId, token string) (auth Auth, err error) {
	authToken := AuthToken{
		RoomID: roomId,
		Token:  token,
		Uid:    uid,
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
