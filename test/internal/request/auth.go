package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	user "gitlab.com/jetfueltw/cpw/alakazam/logic/client"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/permission"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/store"
	"gitlab.com/jetfueltw/cpw/alakazam/pkg/bufio"
	pd "gitlab.com/jetfueltw/cpw/alakazam/protocol"
	"gitlab.com/jetfueltw/cpw/alakazam/protocol/grpc"
	"gitlab.com/jetfueltw/cpw/alakazam/test/internal/protocol"
	"gitlab.com/jetfueltw/cpw/alakazam/test/internal/run"
	"golang.org/x/net/websocket"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"
)

type AuthToken struct {
	Ticket string `json:"ticket"`
	RoomID string `json:"room_id"`
}

type ConnectReply struct {
	RoomId     string                `json:"room_id"`
	Uid        string                `json:"Uid"`
	Key        string                `json:"Key"`
	Permission permission.Permission `json:"permission"`
	Token      string                `json:"token"`
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
	b, _ := uuid.New().MarshalBinary()
	return DialAuthToken(fmt.Sprintf("%x", b), roomId, uuid.New().String())
}

func DialAuthUser(uid, roomId string) (auth Auth, err error) {
	return DialAuthToken(uid, roomId, uuid.New().String())
}

func DialAuthToken(uid, roomId, ticket string) (auth Auth, err error) {
	u := authApi{uid}
	return DialAuthUserByAuthApi(roomId, ticket, u.authApi())
}

func DialAuthUserByAuthApi(roomId, ticket string, authApi run.TransportFunc) (auth Auth, err error) {
	authToken := AuthToken{
		RoomID: roomId,
		Ticket: ticket,
	}
	var (
		conn *websocket.Conn
	)

	run.AddClient("/authentication", authApi)

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

type authApi struct {
	uuid string
}

func (a authApi) authApi() run.TransportFunc {
	return func(request *http.Request) (i *http.Response, e error) {
		u := user.User{
			Uid:      a.uuid,
			Nickname: "test",
			Type:     store.Player,
			Avatar:   "/",
			Token:    "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpYXQiOjE1NTg2ODgwMTcsImlzcyI6ImNwdyIsImF1ZCI6ImNoYXQiLCJzZXNzaW9uX3Rva2VuIjoiY2MwZGEwNjMwMzg2NGFjNWJlZGJhMzViNWQ1NWNkZTEiLCJ1aWQiOiI5ODQxNjQyNmU0OTQ0ZWUyODhkOTQ3NWNkODBiYzUwMSJ9.sfIKY2nZ6b4pWGrAmNUV8ndkQRmnv2fKdg80cW3FS9Y",
		}

		b, err := json.Marshal(u)
		if err != nil {
			return nil, err
		}
		return ToResponse(b)
	}
}

func ToResponse(b []byte) (*http.Response, error) {
	header := http.Header{}
	header.Set("Content-Type", "application/json")

	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(bytes.NewReader(b)),
		Header:     header,
	}, nil
}
