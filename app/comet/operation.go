package comet

import (
	"context"
	"encoding/json"
	"fmt"
	cometpb "gitlab.com/jetfueltw/cpw/alakazam/app/comet/pb"
	logicpb "gitlab.com/jetfueltw/cpw/alakazam/app/logic/pb"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding/gzip"
)

// 告知logic service有人想要進入某個房間
func (s *Server) Connect(c context.Context, p *cometpb.Proto) (*logicpb.ConnectReply, error) {
	return s.logic.Connect(c, &logicpb.ConnectReq{
		Server: s.name,
		Token:  p.Body,
	})
}

// client連線中斷，告知logic service需清理此人的連線資料
func (s *Server) Disconnect(c context.Context, uid, key string) error {
	_, err := s.logic.Disconnect(context.Background(), &logicpb.DisconnectReq{
		Server: s.name,
		Uid:    uid,
		Key:    key,
	})
	return err
}

// 告知logic service要刷新某人的在線狀態(session 過期時間)
func (s *Server) Heartbeat(ctx context.Context, ch *Channel) error {
	_, err := s.logic.Heartbeat(ctx, &logicpb.HeartbeatReq{
		Server: s.name,
		Uid:    ch.Uid,
		Key:    ch.Key,
		Name:   ch.Name,
		RoomId: ch.Room.ID,
	})
	return err
}

// 告知logic service現在房間的在線人數
func (s *Server) RenewOnline(ctx context.Context, serverID string, rommCount map[int32]int32) (allRoom map[int32]int32, err error) {
	reply, err := s.logic.RenewOnline(ctx, &logicpb.OnlineReq{
		Server:    s.name,
		RoomCount: rommCount,
	}, grpc.UseCompressor(gzip.Name))
	if err != nil {
		return
	}
	return reply.AllRoomCount, nil
}

type changeRoom struct {
	RoomId int32 `json:"room_id"`
}

var changeRoomReply = `{"room_id":%d, "status":%t}`

// 處理Proto相關邏輯
func (s *Server) Operate(ctx context.Context, p *cometpb.Proto, ch *Channel, b *Bucket) error {
	switch p.Op {
	// 更換房間
	case cometpb.OpChangeRoom:
		p.Op = cometpb.OpChangeRoomReply
		var r changeRoom

		if err := json.Unmarshal(p.Body, &r); err != nil {
			p.Body = []byte(fmt.Sprintf(changeRoomReply, r.RoomId, false))
			return nil
		}

		if err := b.ChangeRoom(r.RoomId, ch); err != nil {
			p.Body = []byte(fmt.Sprintf(changeRoomReply, r.RoomId, false))
			log.Error("change room", zap.Error(err), zap.Binary("data", p.Body))
			return nil
		}

		room, err := s.logic.ChangeRoom(ctx, &logicpb.ChangeRoomReq{
			Uid:    ch.Uid,
			Key:    ch.Key,
			RoomID: r.RoomId,
		})

		if err != nil {
			p.Body = []byte(fmt.Sprintf(changeRoomReply, r.RoomId, false))
			log.Error("change room for logic", zap.Error(err), zap.Binary("data", p.Body))
			return nil
		}
		if room.HeaderMessage != nil {
			p.Body = []byte(fmt.Sprintf(changeRoomReply, r.RoomId, true))
			ch.protoRing.SetAdv()
			ch.Signal()

			p1, err := ch.protoRing.Set()
			if err != nil {
				log.Error("proto ping set for change room")
				return nil
			}

			p1.Op = cometpb.OpRaw
			p1.Body = room.HeaderMessage
		}
	default:
		// TODO error
		p.Body = nil
	}
	return nil
}
