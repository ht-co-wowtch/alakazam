package front

import (
	"bytes"
	"encoding/json"
	. "github.com/smartystreets/goconvey/convey"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/pkg/encoding/binary"
	pd "gitlab.com/jetfueltw/cpw/alakazam/protocol"
	"gitlab.com/jetfueltw/cpw/alakazam/protocol/grpc"
	"gitlab.com/jetfueltw/cpw/alakazam/test/internal/protocol"
	"gitlab.com/jetfueltw/cpw/alakazam/test/internal/request"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
)

// 進入房間成功
func TestAuth(t *testing.T) {
	roomId := "1"

	Convey("房間列表", t, func() {
		a, err := request.DialAuth(roomId)
		if err != nil {
			t.Fatal(err)
		}

		Reset(func() {
			roomId += "1"
		})

		Convey("當進入房間成功", func() {
			Convey("應回傳房間資訊", func() {
				So(pd.OpAuthReply, ShouldEqual, a.Proto.Op)
				So(a.Reply.Uid, ShouldNotBeEmpty)
				So(a.Reply.Key, ShouldNotBeEmpty)
				So(roomId, ShouldEqual, a.Reply.RoomId)
				So(a.Reply.Permission.Message, ShouldBeTrue)
				So(a.Reply.Permission.SendBonus, ShouldBeTrue)
				So(a.Reply.Permission.GetBonus, ShouldBeTrue)
				So(a.Reply.Permission.SendFollow, ShouldBeTrue)
				So(a.Reply.Permission.GetFollow, ShouldBeTrue)
			})
		})

		Convey("當切換房間", func() {
			proto := new(grpc.Proto)
			proto.Op = pd.OpChangeRoom
			proto.Body = []byte(`{"room_id":"ABC"}`)

			if err = protocol.Write(a.Wr, proto); err != nil {
				t.Fatal(err)
			}
			if err := protocol.Read(a.Rd, a.Proto); err != nil {
				t.Fatal(err)
			}

			So(pd.OpChangeRoomReply, ShouldEqual, a.Proto.Op)
			So(`{"room_id":"ABC"}`, ShouldEqual, string(a.Proto.Body))
		})

		Convey("當進行心跳", func() {
			hbProto := new(grpc.Proto)
			hbProto.Op = pd.OpHeartbeat
			hbProto.Body = nil

			err = protocol.Write(a.Wr, hbProto)
			if err != nil {
				t.Fatal(err)
			}

			err = protocol.Read(a.Rd, a.Proto)
			if err != nil {
				t.Fatal(err)
			}

			Convey("應回傳房間在線人數", func() {
				online := binary.BigEndian.Int32(a.Proto.Body)
				So(pd.OpHeartbeatReply, ShouldEqual, a.Proto.Op)
				So(int32(1), ShouldEqual, online)
			})
		})

		Convey("當不進行心跳", func() {
			Convey("應踢出房間", func() {
				err = protocol.Read(a.Rd, a.Proto)
				So(io.EOF, ShouldEqual, err)
			})
		})
	})

	Convey("當只給房間做webSocket連線但不選擇房間", t, func() {
		ws, err := request.Dial()
		if err != nil {
			t.Fatal(err)
		}

		Convey("應踢出房間", func() {
			b := make([]byte, 100)
			_, err = ws.Read(b)
			So(io.EOF, ShouldEqual, err)
		})
	})

	Convey("當進入房間失敗", t, func() {
		_, err := request.DialAuthUserByAuthApi("1111", "", func(i *http.Request) (response *http.Response, e error) {
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(``))),
			}, nil
		})

		Convey("應踢出房間", func() {
			So(io.EOF, ShouldEqual, err)
		})
	})
}

func TestBlockade(t *testing.T) {
	roomId := "2"

	Convey("會員封鎖", t, func() {
		a, err := request.DialAuth(roomId)
		if err != nil {
			t.Fatal(err)
		}

		Reset(func() {
			roomId += "1"
		})

		Convey("當A會員被封鎖", func() {
			request.SetBlockade(a.Uid, "測試")
			a, _ = request.DialAuthUser(a.Uid, roomId)

			Convey("無法進入房間", func() {
				e := new(errors.Error)
				if err := json.Unmarshal(a.Proto.Body, e); err != nil {
					t.Fatal(err)
				}

				So(10024011, ShouldEqual, e.Code)
				So("您在封鎖状态，无法进入聊天室", ShouldEqual, e.Message)
			})

			Convey("當A會員被解封鎖", func() {
				request.DeleteBlockade(a.Uid)
				b, err := request.DialAuthUser(a.Uid, roomId)
				if err != nil {
					t.Fatal(err)
				}

				So(a.Uid, ShouldEqual, b.Uid)
			})
		})
	})
}
