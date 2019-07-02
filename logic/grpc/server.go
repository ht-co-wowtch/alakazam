package grpc

import (
	"context"
	"gitlab.com/jetfueltw/cpw/alakazam/logic"
	pb "gitlab.com/jetfueltw/cpw/alakazam/protocol/grpc"
	rpc "gitlab.com/jetfueltw/cpw/micro/grpc"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"

	// use gzip decoder
	_ "google.golang.org/grpc/encoding/gzip"
)

// New logic grpc server
func New(c *rpc.Conf, l *logic.Logic) *grpc.Server {
	srv := rpc.New(c)
	pb.RegisterLogicServer(srv, &server{l})
	lis, err := net.Listen(c.Network, c.Addr)
	if err != nil {
		panic(err)
	}
	go func() {
		if err := srv.Serve(lis); err != nil {
			panic(err)
		}
	}()
	return srv
}

type server struct {
	srv *logic.Logic
}

var _ pb.LogicServer = &server{}

// Ping Service
func (s *server) Ping(ctx context.Context, req *pb.PingReq) (*pb.PingReply, error) {
	return &pb.PingReply{}, nil
}

// 某client要做連線
func (s *server) Connect(ctx context.Context, req *pb.ConnectReq) (*pb.ConnectReply, error) {
	r, err := s.srv.Connect(req.Server, req.Token)
	if err != nil {
		log.Error("grpc connect", zap.Error(err), zap.String("data", string(req.Token)))
		return &pb.ConnectReply{}, err
	}
	return &pb.ConnectReply{
		Uid:       r.Uid,
		Key:       r.Key,
		Name:      r.Name,
		RoomID:    r.RoomId,
		Heartbeat: r.Hb,
		Status:    int32(r.Permission),
	}, nil
}

// 某client要中斷連線
func (s *server) Disconnect(ctx context.Context, req *pb.DisconnectReq) (*pb.DisconnectReply, error) {
	has, err := s.srv.Disconnect(req.Uid, req.Key, req.Server)
	if err != nil {
		return &pb.DisconnectReply{}, err
	}
	return &pb.DisconnectReply{Has: has}, nil
}

// user當前連線要切換房間
func (s *server) ChangeRoom(ctx context.Context, req *pb.ChangeRoomReq) (*pb.ChangeRoomReply, error) {
	return &pb.ChangeRoomReply{}, s.srv.ChangeRoom(req.Uid, req.Key, req.RoomID)
}

// 重置user redis過期時間
func (s *server) Heartbeat(ctx context.Context, req *pb.HeartbeatReq) (*pb.HeartbeatReply, error) {
	if err := s.srv.Heartbeat(req.Uid, req.Key, req.RoomId, req.Name, req.Server); err != nil {
		return &pb.HeartbeatReply{}, err
	}
	return &pb.HeartbeatReply{}, nil
}

// 更新每個房間線上總人數資料
func (s *server) RenewOnline(ctx context.Context, req *pb.OnlineReq) (*pb.OnlineReply, error) {
	allRoomCount, err := s.srv.RenewOnline(req.Server, req.RoomCount)
	if err != nil {
		return &pb.OnlineReply{}, err
	}
	return &pb.OnlineReply{AllRoomCount: allRoomCount}, nil
}
