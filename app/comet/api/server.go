package api

import (
	"context"
	"net"
	"time"

	"gitlab.com/jetfueltw/cpw/alakazam/app/comet"
	"gitlab.com/jetfueltw/cpw/alakazam/app/comet/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/app/comet/pb"
	rpc "gitlab.com/jetfueltw/cpw/micro/grpc"

	"google.golang.org/grpc"
)

// New comet grpc server.
func New(c *rpc.Conf, s *comet.Server) *grpc.Server {
	srv := rpc.New(c)
	pb.RegisterCometServer(srv, &server{s})
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
	srv *comet.Server
}

var _ pb.CometServer = &server{}

// Ping Service
func (s *server) Ping(ctx context.Context, req *pb.Empty) (*pb.Empty, error) {
	return &pb.Empty{}, nil
}

// 踢人
func (s *server) Kick(ctx context.Context, req *pb.KeyReq) (*pb.Empty, error) {
	for _, key := range req.Key {
		if b := s.srv.Bucket(key); b != nil {
			if ch := b.Channel(key); ch != nil {
				_ = ch.Push(req.Proto)
				ch.Close()
			}
		}
	}
	return &pb.Empty{}, nil
}

// 所有房間做推送
func (s *server) Broadcast(ctx context.Context, req *pb.BroadcastReq) (*pb.BroadcastReply, error) {
	if req.Proto == nil {
		return nil, errors.ErrBroadCastArg
	}
	// TODO use broadcast queue
	go func() {
		for _, bucket := range s.srv.Buckets() {
			bucket.Broadcast(req.GetProto())
			if req.Speed > 0 {
				t := bucket.ChannelCount() / int(req.Speed)
				time.Sleep(time.Duration(t) * time.Second)
			}
		}
	}()
	return &pb.BroadcastReply{}, nil
}

// 單一房間推送
func (s *server) BroadcastRoom(ctx context.Context, req *pb.BroadcastRoomReq) (*pb.BroadcastRoomReply, error) {
	if req.Proto == nil || req.RoomID == 0 {
		return nil, errors.ErrBroadCastRoomArg
	}
	for _, bucket := range s.srv.Buckets() {
		bucket.BroadcastRoom(req)
	}
	return &pb.BroadcastRoomReply{}, nil
}

// server上有哪些房間
func (s *server) Rooms(ctx context.Context, req *pb.RoomsReq) (*pb.RoomsReply, error) {
	var (
		roomIds = make(map[int32]bool)
	)
	for _, bucket := range s.srv.Buckets() {
		for roomID := range bucket.Rooms() {
			roomIds[roomID] = true
		}
	}
	return &pb.RoomsReply{Rooms: roomIds}, nil
}
