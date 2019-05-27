package comet

import (
	"context"
	log "github.com/golang/glog"
	"gitlab.com/jetfueltw/cpw/alakazam/protocol"
	pd "gitlab.com/jetfueltw/cpw/alakazam/protocol/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding/gzip"
)

// 告知logic service有人想要進入某個房間
func (s *Server) Connect(c context.Context, p *pd.Proto) (*pd.ConnectReply, error) {
	return s.rpcClient.Connect(c, &pd.ConnectReq{
		Server: s.name,
		Token:  p.Body,
	})
}

// client連線中斷，告知logic service需清理此人的連線資料
func (s *Server) Disconnect(c context.Context, uid, key string) (err error) {
	_, err = s.rpcClient.Disconnect(context.Background(), &pd.DisconnectReq{
		Server: s.name,
		Uid:    uid,
		Key:    key,
	})
	return
}

// 告知logic service要刷新某人的在線狀態(session 過期時間)
func (s *Server) Heartbeat(ctx context.Context, ch *Channel) (err error) {
	_, err = s.rpcClient.Heartbeat(ctx, &pd.HeartbeatReq{
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
	reply, err := s.rpcClient.RenewOnline(ctx, &pd.OnlineReq{
		Server:    s.name,
		RoomCount: rommCount,
	}, grpc.UseCompressor(gzip.Name))
	if err != nil {
		return
	}
	return reply.AllRoomCount, nil
}

// 處理Proto相關邏輯
func (s *Server) Operate(ctx context.Context, p *pd.Proto, ch *Channel, b *Bucket) error {
	switch p.Op {
	// 更換房間
	case protocol.OpChangeRoom:
		if err := b.ChangeRoom(string(p.Body), ch); err != nil {
			log.Errorf("Change Room (%s) error(%v)", p.Body, err)
		} else if _, err := s.rpcClient.ChangeRoom(ctx, &pd.ChangeRoomReq{
			Uid:    ch.Uid,
			Key:    ch.Key,
			RoomID: string(p.Body),
		}); err != nil {
			return err
		}
		p.Op = protocol.OpChangeRoomReply
	default:
		// TODO error
		p.Body = nil
	}
	return nil
}
