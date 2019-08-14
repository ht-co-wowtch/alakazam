package comet

import (
	"context"
	"encoding/json"
	cometpb "gitlab.com/jetfueltw/cpw/alakazam/comet/pb"
	logicpb "gitlab.com/jetfueltw/cpw/alakazam/logic/pb"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding/gzip"
)

// 告知logic service有人想要進入某個房間
func (s *Server) Connect(c context.Context, p *cometpb.Proto) (*logicpb.ConnectReply, error) {
	return s.rpcClient.Connect(c, &logicpb.ConnectReq{
		Server: s.name,
		Token:  p.Body,
	})
}

// client連線中斷，告知logic service需清理此人的連線資料
func (s *Server) Disconnect(c context.Context, uid, key string) (err error) {
	_, err = s.rpcClient.Disconnect(context.Background(), &logicpb.DisconnectReq{
		Server: s.name,
		Uid:    uid,
		Key:    key,
	})
	return
}

// 告知logic service要刷新某人的在線狀態(session 過期時間)
func (s *Server) Heartbeat(ctx context.Context, ch *Channel) (err error) {
	_, err = s.rpcClient.Heartbeat(ctx, &logicpb.HeartbeatReq{
		Server: s.name,
		Uid:    ch.Uid,
		Key:    ch.Key,
		Name:   ch.Name,
		RoomId: ch.Room.ID,
	})
	return
}

// 告知logic service現在房間的在線人數
func (s *Server) RenewOnline(ctx context.Context, serverID string, rommCount map[string]int32) (allRoom map[string]int32, err error) {
	reply, err := s.rpcClient.RenewOnline(ctx, &logicpb.OnlineReq{
		Server:    s.name,
		RoomCount: rommCount,
	}, grpc.UseCompressor(gzip.Name))
	if err != nil {
		return
	}
	return reply.AllRoomCount, nil
}

type changeRoom struct {
	RoomId string `json:"room_id"`
}

// 處理Proto相關邏輯
func (s *Server) Operate(ctx context.Context, p *cometpb.Proto, ch *Channel, b *Bucket) error {
	switch p.Op {
	// 更換房間
	case cometpb.OpChangeRoom:
		var r changeRoom

		if err := json.Unmarshal(p.Body, &r); err != nil {
			return err
		}

		if err := b.ChangeRoom(r.RoomId, ch); err != nil {
			log.Error("change room", zap.Error(err), zap.Binary("data", p.Body))
		} else if _, err := s.rpcClient.ChangeRoom(ctx, &logicpb.ChangeRoomReq{
			Uid:    ch.Uid,
			Key:    ch.Key,
			RoomID: r.RoomId,
		}); err != nil {
			return err
		}

		p.Op = cometpb.OpChangeRoomReply
	default:
		// TODO error
		p.Body = nil
	}
	return nil
}
