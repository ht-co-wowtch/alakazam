package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/api/comet/grpc"
	"gitlab.com/jetfueltw/cpw/alakazam/protocol"
	"gitlab.com/jetfueltw/cpw/alakazam/pkg/bufio"
	"gitlab.com/jetfueltw/cpw/alakazam/pkg/encoding/binary"
	"golang.org/x/net/websocket"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"testing"
	"time"
)

const (
	// Protocol 長度的byte長度
	_packSize = 4

	// Protocol Header的byte長度
	_headerSize = 2

	// Protocol 版本號的byte長度
	_verSize = 2

	// Protocol 動作意義的byte長度
	_opSize = 4

	// Protocol seq的byte長度
	_seqSize = 4

	// Protocol Header的總長度
	_rawHeaderSize = _packSize + _headerSize + _verSize + _opSize + _seqSize

	// Protocol 長度的byte位置範圍
	_packOffset = 0

	// Protocol 整個header長度的byte位置範圍
	// Protocol 長度 - header長度 = Body長度
	_headerOffset = _packOffset + _packSize

	// Protocol版本號的byte位置範圍
	_verOffset = _headerOffset + _headerSize

	// Protocol動作意義的byte位置範圍
	_opOffset = _verOffset + _verSize

	// Protocol seq意義的byte位置範圍
	_seqOffset = _opOffset + _opSize

	host = "http://127.0.0.1:3111"
)

type AuthToken struct {
	Mid      int64   `json:"mid"`
	Key      string  `json:"key"`
	RoomID   string  `json:"room_id"`
	Platform string  `json:"platform"`
	Accepts  []int32 `json:"accepts"`
}

type auth struct {
	wr    *bufio.Writer
	rd    *bufio.Reader
	proto *grpc.Proto
}

var (
	authToken  *AuthToken
	httpClient *http.Client
)

func init() {
	rand.Seed(time.Now().Unix())
	authToken = &AuthToken{
		0,
		"",
		"chat://1000",
		"web",
		[]int32{1, 1000},
	}

	httpClient = &http.Client{
		Timeout: time.Second * 5,
	}
}

func Test_auth(t *testing.T) {
	a, err := dialAuth(authToken)
	if err != nil {
		assert.Error(t, err)
		return
	}
	shouldBeAuthReply(t, a)
}

func Test_not_auth(t *testing.T) {
	ws, err := dial()
	if err != nil {
		assert.Error(t, err)
		return
	}
	shouldBeCloseConnection(err, ws, t)
}

func Test_heartbeat(t *testing.T) {
	a, err := dialAuth(authToken)
	if err != nil {
		assert.Error(t, err)
		return
	}
	shouldBeHeartbeatReply(t, a, givenHeartbeat())
}

func Test_not_heartbeat(t *testing.T) {
	a, err := dialAuth(authToken)
	if err != nil {
		assert.Error(t, err)
		return
	}
	shouldBeTimeoutConnection(err, a, t)
}

func Test_push_user(t *testing.T) {
	pushTest(t, authToken, func() ([]byte, error) {
		return pushUser(authToken.Mid, "測試")
	}, func(p []grpc.Proto, otherErr error, otherProto []grpc.Proto) {
		assert.Equal(t, []byte(`測試`), p[0].Body)
		assert.Nil(t, otherErr)
		assert.Len(t, otherProto, 0)
	})
}

func Test_push_room(t *testing.T) {
	pushTest(t, authToken, func() ([]byte, error) {
		return pushRoom(1000, "測試")
	}, func(p []grpc.Proto, otherErr error, otherProto []grpc.Proto) {
		assert.Equal(t, []byte(`測試`), p[0].Body)
		assert.Nil(t, otherErr)
		assert.Equal(t, []byte(`測試`), otherProto[0].Body)
	})
}

func Test_push_broadcast(t *testing.T) {
	other := *authToken
	other.RoomID = "chat://1001"
	pushTest(t, &other, func() ([]byte, error) {
		return pushBroadcast("測試")
	}, func(p []grpc.Proto, otherErr error, otherProto []grpc.Proto) {
		assert.Equal(t, []byte(`測試`), p[0].Body)
		assert.Nil(t, otherErr)
		assert.Equal(t, []byte(`測試`), otherProto[0].Body)
	})
}

func pushTest(t *testing.T, otherAuth *AuthToken, f func() ([]byte, error), ass func(p []grpc.Proto, otherErr error, otherProto []grpc.Proto)) {
	a, err := dialAuth(authToken)
	if err != nil {
		assert.Error(t, err)
		return
	}

	var (
		other      auth
		otherErr   error
		otherProto []grpc.Proto
	)

	go func() {
		other, otherErr = dialAuth(otherAuth)
		otherProto, otherErr = readMessageProto(other.rd, other.proto)
	}()

	b, err := f()
	if err != nil {
		assert.Error(t, err)
		return
	}
	time.Sleep(time.Second * 3)
	var p []grpc.Proto
	if p, err = readMessageProto(a.rd, a.proto); err != nil {
		assert.Error(t, err)
		return
	}

	assert.Equal(t, protocol.OpRaw, a.proto.Op)
	assert.Equal(t, []byte(`{"code":0,"message":""}`), b)
	ass(p, otherErr, otherProto)
	fmt.Println("ok")
}

func shouldBeTimeoutConnection(err error, a auth, t *testing.T) {
	fmt.Println(time.Now())
	err = readProto(a.rd, a.proto)
	fmt.Println(time.Now())
	assert.Equal(t, io.EOF, err)
	fmt.Println("ok")
}

func shouldBeCloseConnection(err error, ws *websocket.Conn, t *testing.T) {
	b := make([]byte, 100)
	_, err = ws.Read(b)
	assert.Equal(t, io.EOF, err)
	fmt.Println("ok")
}

func givenHeartbeat() *grpc.Proto {
	hbProto := new(grpc.Proto)
	hbProto.Op = protocol.OpHeartbeat
	hbProto.Seq = 1
	hbProto.Body = nil
	return hbProto
}

func shouldBeAuthReply(t *testing.T, a auth) {
	assert.Equal(t, protocol.OpAuthReply, a.proto.Op)
	fmt.Println("ok")
}

func shouldBeHeartbeatReply(t *testing.T, a auth, hbProto *grpc.Proto) {
	fmt.Println("send heartbeat")
	if err := writeProto(a.wr, hbProto); err != nil {
		assert.Error(t, err)
		return
	}
	if err := readProto(a.rd, a.proto); err != nil {
		assert.Error(t, err)
		return
	}
	fmt.Println("heartbeat reply")
	assert.Equal(t, protocol.OpHeartbeatReply, a.proto.Op)
	fmt.Println("ok")
}

func dial() (conn *websocket.Conn, err error) {
	conn, err = websocket.Dial("ws://127.0.0.1:3102/sub", "", "http://127.0.0.1")

	return
}

func dialAuth(authToken *AuthToken) (auth auth, err error) {
	authToken.Mid = rand.Int63()
	var (
		conn *websocket.Conn
	)

	conn, err = dial()

	wr := bufio.NewWriter(conn)
	rd := bufio.NewReader(conn)

	proto := new(grpc.Proto)
	proto.Ver = 1
	proto.Op = protocol.OpAuth
	proto.Seq = int32(0)
	proto.Body, _ = json.Marshal(authToken)

	fmt.Printf("send auth: %s\n", proto.Body)
	if err = writeProto(wr, proto); err != nil {
		return
	}
	if err = readProto(rd, proto); err != nil {
		return
	}
	fmt.Println("auth reply")

	auth.wr = wr
	auth.rd = rd
	auth.proto = proto
	return
}

func writeProto(wr *bufio.Writer, p *grpc.Proto) (err error) {
	var (
		buf     []byte
		packLen int32
	)

	packLen = _rawHeaderSize + int32(len(p.Body))
	if buf, err = wr.Peek(_rawHeaderSize); err != nil {
		return
	}
	binary.BigEndian.PutInt32(buf[_packOffset:], packLen)
	binary.BigEndian.PutInt16(buf[_headerOffset:], int16(_rawHeaderSize))
	binary.BigEndian.PutInt16(buf[_verOffset:], int16(p.Ver))
	binary.BigEndian.PutInt32(buf[_opOffset:], p.Op)
	binary.BigEndian.PutInt32(buf[_seqOffset:], p.Seq)
	if p.Body != nil {
		_, err = wr.Write(p.Body)
	}
	return wr.Flush()
}

func readProto(rr *bufio.Reader, p *grpc.Proto) (err error) {
	var (
		bodyLen   int
		headerLen int16
		packLen   int32
	)

	packLen, headerLen, err = read(rr, p)
	if bodyLen = int(packLen - int32(headerLen)); bodyLen > 0 {
		p.Body, err = rr.Pop(bodyLen)
	} else {
		p.Body = nil
	}
	return
}

func readMessageProto(rr *bufio.Reader, p *grpc.Proto) (protos []grpc.Proto, err error) {
	var (
		bodyLen   int
		headerLen int16
		packLen   int32
	)

	packLen, headerLen, err = read(rr, p)
	for offset := int(headerLen); offset < int(packLen); offset += int(packLen) {
		proto := new(grpc.Proto)
		packLen, headerLen, err = read(rr, proto)
		if bodyLen = int(packLen - int32(headerLen)); bodyLen > 0 {
			proto.Body, err = rr.Pop(bodyLen)
		} else {
			proto.Body = nil
		}
		protos = append(protos, *proto)
	}
	return
}

func read(rr *bufio.Reader, p *grpc.Proto) (packLen int32, headerLen int16, err error) {
	var (
		buf []byte
	)
	if buf, err = rr.Pop(_rawHeaderSize); err != nil {
		return
	}

	packLen = binary.BigEndian.Int32(buf[_packOffset:_headerOffset])
	headerLen = binary.BigEndian.Int16(buf[_headerOffset:_verOffset])
	p.Ver = int32(binary.BigEndian.Int16(buf[_verOffset:_opOffset]))
	p.Op = binary.BigEndian.Int32(buf[_opOffset:_seqOffset])
	p.Seq = binary.BigEndian.Int32(buf[_seqOffset:])
	return
}

func pushUser(id int64, message string) ([]byte, error) {
	return push(fmt.Sprintf(host+"/goim/push/mids?operation=1000&mids=%d", id), bytes.NewBufferString(message))
}

func pushRoom(roomId int, message string) ([]byte, error) {
	return push(fmt.Sprintf(host+"/goim/push/room?operation=1000&type=chat&room=%d", roomId), bytes.NewBufferString(message))
}

func pushBroadcast(message string) ([]byte, error) {
	return push(fmt.Sprintf(host+"/goim/push/all?operation=1000"), bytes.NewBufferString(message))
}

func push(url string, message io.Reader) (body []byte, err error) {
	resp, err := httpPost(url, "", message)
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

func httpPost(url string, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
