package rpc

import (
	"context"
	"gitlab.com/jetfueltw/cpw/micro/id"
	"time"

	"gitlab.com/jetfueltw/cpw/alakazam/app/logic/pb"
	"gitlab.com/jetfueltw/cpw/alakazam/client"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/message"
	"gitlab.com/jetfueltw/cpw/alakazam/room"
	"gitlab.com/jetfueltw/cpw/micro/errdefs"
	rpc "gitlab.com/jetfueltw/cpw/micro/grpc"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	// use gzip decoder
	_ "google.golang.org/grpc/encoding/gzip"
)

// New logic grpc server
func New(c *rpc.Conf, room room.Chat, message *message.Producer, cli *client.Client) *grpc.Server {
	srv := rpc.New(c)
	pb.RegisterLogicServer(srv, &server{
		room:    room,
		message: message,
		cli:     cli,
	})

	return srv
}

type server struct {
	room    room.Chat
	message *message.Producer
	cli     *client.Client
}

var _ pb.LogicServer = &server{}

// Ping Service
func (s *server) Ping(ctx context.Context, req *pb.PingReq) (*pb.PingReply, error) {
	return &pb.PingReply{}, nil
}

// 某client要做連線
func (s *server) Connect(ctx context.Context, req *pb.ConnectReq) (*pb.ConnectReply, error) {
	connect, err := s.room.Connect(req.Server, req.Token)
	if err != nil {
		log.Error("[rpc/server.go]grpc connect", zap.Error(err), zap.String("data", string(req.Token)))
		switch e := err.(type) {
		case errdefs.Error:
			return &pb.ConnectReply{}, status.Error(codes.FailedPrecondition, err.Error())
		case *errdefs.Causer:
			var msg string
			if e.Code == errors.NoLogin {
				msg = errors.NoLoginMessage
			} else {
				msg = e.Message
			}
			return &pb.ConnectReply{}, status.Error(codes.FailedPrecondition, msg)
		}
		return &pb.ConnectReply{}, status.Error(codes.Internal, err.Error())
	}
	return connect, nil
}

// 成功進入房間
func (s *server) ConnectSuccessReply(ctx context.Context, req *pb.ConnectSuccessReq) (*pb.PingReply, error) {
	// 主播不顯示進場訊息
	if req.User.Type == 3 {
		return &pb.PingReply{}, nil
	}

	_, err := s.message.SendConnect(req.RoomId, req.User, req.IsManage)
	return &pb.PingReply{}, err
}

// 某client要中斷連線
func (s *server) Disconnect(ctx context.Context, req *pb.DisconnectReq) (*pb.DisconnectReply, error) {
	has, err := s.room.Disconnect(req.Uid, req.Key)
	if err != nil {
		log.Error("[rpc/server.go]Disconnect", zap.Error(err), zap.String("uid", req.Uid))
		return &pb.DisconnectReply{}, err
	} else {
		log.Info("[rpc/server.go]Disconnect", zap.String("uid", req.Uid), zap.String("key", req.Key))
	}
	return &pb.DisconnectReply{Has: has}, nil
}

// user當前連線要切換房間
func (s *server) ChangeRoom(ctx context.Context, req *pb.ChangeRoomReq) (*pb.ConnectReply, error) {
	p, err := s.room.ChangeRoom(req.Uid, int(req.RoomID), req.Key)
	if err != nil {
		log.Error("[rpc/server.go]ChangeRoom", zap.Error(err), zap.Int32("rid", req.RoomID))
		switch e := err.(type) {
		case errdefs.Error:
			return &pb.ConnectReply{}, status.Error(codes.FailedPrecondition, err.Error())
		case *errdefs.Causer:
			var msg string
			if e.Code == errors.NoLogin {
				msg = errors.NoLoginMessage
			} else {
				msg = e.Message
			}
			return &pb.ConnectReply{}, status.Error(codes.FailedPrecondition, msg)
		}
		return &pb.ConnectReply{}, status.Error(codes.Internal, err.Error())
	}
	return p, nil
}

// 重置user redis過期時間
func (s *server) Heartbeat(ctx context.Context, req *pb.HeartbeatReq) (*pb.HeartbeatReply, error) {
	if err := s.room.Heartbeat(req.Uid, req.Key, req.Name, req.Server); err != nil {
		log.Error("[rpc/server.go]Heartbeat", zap.Error(err), zap.String("uid", req.Uid), zap.Int32("room_id", req.RoomId))
		return &pb.HeartbeatReply{}, err
	}
	return &pb.HeartbeatReply{}, nil
}

// 更新每個房間線上總人數資料
func (s *server) RenewOnline(ctx context.Context, req *pb.OnlineReq) (*pb.OnlineReply, error) {
	allRoomCount, err := s.room.RenewOnline(req.Server, req.RoomCount)
	if err != nil {
		log.Error("[rpc/server.go]RenewOnline", zap.Error(err), zap.String("server", req.Server))
		return &pb.OnlineReply{}, err
	}
	return &pb.OnlineReply{
		AllRoomCount: allRoomCount,
	}, nil
}

// 付費房月卡效期
// PaidRoomExpiry
func (s *server) PaidRoomExpiry(ctx context.Context, req *pb.MemberProfileReq) (*pb.MemberProfileReply, error) {
	resp, _ := s.cli.LiveExpire(req.Uid)

	// 時間檢查
	isAllow := false
	if time.Now().Before(resp.LiveExpireAt) {
		isAllow = true
	}

	return &pb.MemberProfileReply{
		Expire:  resp.LiveExpireAt.String(),
		IsAllow: isAllow,
	}, nil
}

// 付費房鑽石付費
// PaidRoomDiamond
func (s *server) PaidRoomDiamond(ctx context.Context, req *pb.PaidRoomDiamondReq) (*pb.PaidRoomDiamondReply, error) {
	// 取得收費房收費標準
	lr, err := s.cli.GetLiveChatInfo(req.RoomID)
	if err != nil {
		log.Infof("GetLiveChatInfo error, %o", err)
		return nil, err
	}
	if !lr.IsLive {
		log.Errorf("lr %o", lr)
		return &pb.PaidRoomDiamondReply{
			Status: false,
			Error:  "关播中",
		}, nil
	}

	uid := id.UUid32()

	// 跨帳鑽石異動
	tr, err := s.cli.PaidDiamond(client.PaidDiamondTXTOrder{ // TODO
		From: client.PaidDiamondUser{Uid: req.Uid, Type: "diamond-sub"},
		To:   client.PaidDiamondUser{Uid: lr.MemberUid, Type: "diamond-add"},
		Orders: []client.PaidDiamondOrder{
			{
				Id:     uid,
				Amount: -lr.Charge,
			},
		},
	})

	if err != nil {
		return &pb.PaidRoomDiamondReply{
			Status: false,
			Error:  err.Error(),
		}, nil
	}

	paidTime := time.Now()

	// 建立鑽石付費訂單
	_, err = s.cli.CreateLiveChatPaidOrder(lr.SiteId, req.Uid, lr.Id, uid, lr.Charge)

	if err != nil {
		log.Infof("CreateLiveChatPaidOrder error, %o", err)
		return &pb.PaidRoomDiamondReply{}, nil // TODO
	}

	return &pb.PaidRoomDiamondReply{
		Status:   true,
		Diamond:  tr.From.Diamond,
		PaidTime: paidTime.String(),
	}, nil
}
