package request

import (
	"encoding/json"
	"fmt"
	"gitlab.com/jetfueltw/cpw/alakazam/comet/pb"
	"gitlab.com/jetfueltw/cpw/alakazam/pkg/bufio"
	"gitlab.com/jetfueltw/cpw/alakazam/test/internal/protocol"
	"gitlab.com/jetfueltw/cpw/micro/client"
	"gitlab.com/jetfueltw/cpw/micro/errdefs"
	"golang.org/x/net/websocket"
	"math/rand"
	"net/http"
	"testing"
	"time"
)

type authToken struct {
	Token  string `json:"token"`
	RoomID string `json:"room_id"`
}

type ConnectReply struct {
	RoomId     string `json:"room_id"`
	Uid        string `json:"Uid"`
	Key        string `json:"Key"`
	Permission struct {
		IsBanned      bool `json:"is_banned"`
		IsRedEnvelope bool `json:"is_red_envelope"`
	} `json:"permission"`
}

type Auth struct {
	*ConnectReply

	Wr    *bufio.Writer
	Rd    *bufio.Reader
	Proto *pb.Proto
}

var (
	token *client.Client
)

func init() {
	rand.Seed(time.Now().Unix())

	token = client.New(&client.Conf{
		Host:            "127.0.0.1:9000",
		Scheme:          "http",
		MaxConns:        5,
		MaxIdleConns:    1,
		IdleConnTimeout: time.Second * 2,
	})
}

func Dial() (conn *websocket.Conn, err error) {
	conn, err = websocket.Dial("ws://127.0.0.1:3102/sub", "", "http://127.0.0.1")
	return
}

func DialAuth(t *testing.T, roomId, uid string) *Auth {
	auth, err := DialAuthToken(roomId, GetToken(t, uid))
	if err != nil {
		t.Fatal(err)
	}
	return auth
}

func DialAuthToken(roomId, token string) (*Auth, error) {
	authToken := authToken{
		RoomID: roomId,
		Token:  token,
	}
	var (
		conn *websocket.Conn
		err  error
	)

	conn, err = Dial()
	if err != nil {
		return nil, err
	}

	wr := bufio.NewWriter(conn)
	rd := bufio.NewReader(conn)

	proto := new(pb.Proto)
	proto.Op = pb.OpAuth
	proto.Body, _ = json.Marshal(authToken)

	if err = protocol.Write(wr, proto); err != nil {
		return nil, err
	}
	if err = protocol.Read(rd, proto); err != nil {
		return nil, err
	}

	auth := new(Auth)
	auth.Wr = wr
	auth.Rd = rd
	auth.Proto = proto

	reply := new(ConnectReply)

	if err = json.Unmarshal(proto.Body, &reply); err != nil {
		return nil, err
	}

	auth.ConnectReply = reply
	return auth, nil
}

func (a *Auth) ChangeRoom(roomId string) error {
	proto := new(pb.Proto)
	proto.Op = pb.OpChangeRoom
	proto.Body = []byte(fmt.Sprintf(`{"room_id":"%s"}`, roomId))
	if err := protocol.Write(a.Wr, proto); err != nil {
		return err
	}
	if err := protocol.Read(a.Rd, a.Proto); err != nil {
		return err
	}
	return nil
}

func (a *Auth) PushRoom(message string) Response {
	return PushRoom(a.Uid, a.Key, message)
}

func (a *Auth) Heartbeat() error {
	hbProto := new(pb.Proto)
	hbProto.Op = pb.OpHeartbeat
	hbProto.Body = nil
	if err := protocol.Write(a.Wr, hbProto); err != nil {
		return err
	}
	if err := protocol.Read(a.Rd, a.Proto); err != nil {
		return err
	}
	return nil
}

func (a *Auth) Read() error {
	return protocol.Read(a.Rd, a.Proto)
}

func (a *Auth) ReadMessage() ([]pb.Proto, error) {
	return protocol.ReadMessage(a.Rd, a.Proto)
}

func (a *Auth) SetBlockade(remark string) Response {
	return SetBlockade(a.Uid, remark)
}

func (a *Auth) DeleteBlockade(remark string) Response {
	return DeleteBlockade(a.Uid)
}

func (a *Auth) SetBanned(remark string, sec int) Response {
	return SetBanned(a.Uid, remark, sec)
}

func (a *Auth) DeleteBanned() Response {
	return DeleteBanned(a.Uid)
}

func GetToken(t *testing.T, uid string) string {
	token, err := getUserToken(uid)
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func getUserToken(uid string) (string, error) {
	var body struct {
		Uid             string                 `json:"uid"`
		SiteCode        string                 `json:"site_code"`
		DelOtherSession bool                   `json:"del_other_session"`
		SessionData     map[string]interface{} `json:"session_data"`
	}
	body.Uid = uid
	body.SiteCode = "default"
	body.SessionData = map[string]interface{}{
		"id":       1,
		"username": "sam78",
		"type":     2,
	}

	resp, err := token.PostJson("/sessions", nil, body, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if err := errResponse(resp); err != nil {
		return "", err
	}

	var token struct {
		SessionToken string `json:"session_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return "", err
	}
	return token.SessionToken, nil
}

func errResponse(resp *http.Response) *errdefs.Error {
	if resp.StatusCode != http.StatusOK {
		e := new(errdefs.Error)
		if err := json.NewDecoder(resp.Body).Decode(e); err != nil {
			return &errdefs.Error{Err: err}
		}
		return e
	}
	return nil
}
