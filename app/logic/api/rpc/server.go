package rpc

import (
	"context"
	"gitlab.com/jetfueltw/cpw/alakazam/app/logic/pb"
	"gitlab.com/jetfueltw/cpw/alakazam/room"
	"gitlab.com/jetfueltw/cpw/micro/errdefs"
	rpc "gitlab.com/jetfueltw/cpw/micro/grpc"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	// use gzip decoder
	_ "google.golang.org/grpc/encoding/gzip"
)

// New logic grpc server
func New(c *rpc.Conf, room *room.Room) *grpc.Server {
	srv := rpc.New(c)
	pb.RegisterLogicServer(srv, &server{room})
	return srv
}

type server struct {
	room *room.Room
}

var _ pb.LogicServer = &server{}

// Ping Service
func (s *server) Ping(ctx context.Context, req *pb.PingReq) (*pb.PingReply, error) {
	return &pb.PingReply{}, nil
}

// 某client要做連線
func (s *server) Connect(ctx context.Context, req *pb.ConnectReq) (*pb.ConnectReply, error) {
	r, err := s.room.Connect(req.Server, req.Token)
	if err != nil {
		if _, ok := err.(*errdefs.Error); !ok {
			log.Error("grpc connect", zap.Error(err), zap.String("data", string(req.Token)))
		}
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
	has, err := s.room.Disconnect(req.Uid, req.Key)
	if err != nil {
		log.Error("grpc disconnect", zap.Error(err), zap.String("uid", req.Uid))
		return &pb.DisconnectReply{}, err
	} else {
		log.Info("conn disconnect", zap.String("uid", req.Uid), zap.String("key", req.Key))
	}
	return &pb.DisconnectReply{Has: has}, nil
}

// user當前連線要切換房間
func (s *server) ChangeRoom(ctx context.Context, req *pb.ChangeRoomReq) (*pb.ChangeRoomReply, error) {
	return &pb.ChangeRoomReply{}, nil
}

// 重置user redis過期時間
func (s *server) Heartbeat(ctx context.Context, req *pb.HeartbeatReq) (*pb.HeartbeatReply, error) {
	if err := s.room.Heartbeat(req.Uid, req.Key, req.RoomId, req.Name, req.Server); err != nil {
		log.Error("grpc heart beat", zap.Error(err), zap.String("uid", req.Uid), zap.String("room_id", req.RoomId))
		return &pb.HeartbeatReply{}, err
	}
	return &pb.HeartbeatReply{}, nil
}

// 更新每個房間線上總人數資料
func (s *server) RenewOnline(ctx context.Context, req *pb.OnlineReq) (*pb.OnlineReply, error) {
	allRoomCount, err := s.room.RenewOnline(req.Server, req.RoomCount)
	if err != nil {
		log.Error("grpc renew online", zap.Error(err), zap.String("server", req.Server))
		return &pb.OnlineReply{}, err
	}
	return &pb.OnlineReply{AllRoomCount: allRoomCount}, nil
}
