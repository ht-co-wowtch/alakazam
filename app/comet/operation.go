package comet

import (
	"context"
	"encoding/json"

	cometpb "gitlab.com/jetfueltw/cpw/alakazam/app/comet/pb"
	logicpb "gitlab.com/jetfueltw/cpw/alakazam/app/logic/pb"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/status"
)

// 告知logic service有人想要進入某個房間
func (s *Server) Connect(c context.Context, p *cometpb.Proto) (*logicpb.ConnectReply, error) {
	return s.logic.Connect(c, &logicpb.ConnectReq{
		Server: s.name,
		Token:  p.Body,
	})
}

// 進入某個房間成功回應
func (s *Server) ConnectSuccessReply(c context.Context, rid int32, user *logicpb.User, connect *logicpb.Connect) (*logicpb.PingReply, error) {
	return s.logic.ConnectSuccessReply(c, &logicpb.ConnectSuccessReq{
		RoomId:   rid,
		User:     user,
		IsManage: connect.Permission.IsManage,
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

// 處理Proto相關邏輯
func (s *Server) Operate(ctx context.Context, p *cometpb.Proto, ch *Channel, b *Bucket) error {
	switch p.Op {
	// 更換房間
	case cometpb.OpChangeRoom:
		p.Op = cometpb.OpChangeRoomReply
		var r changeRoom

		if err := json.Unmarshal(p.Body, &r); err != nil {
			re := newConnect()
			re.Message = "切换房间失败"
			p.Body, _ = json.Marshal(re)
			return nil
		}

		reply, err := s.logic.ChangeRoom(ctx, &logicpb.ChangeRoomReq{
			Uid:    ch.Uid,
			Key:    ch.Key,
			RoomID: r.RoomId,
		})

		if err != nil {
			re := newConnect()
			s, _ := status.FromError(err)
			if s.Code() != codes.FailedPrecondition {
				log.Error("change room for logic", zap.Error(err), zap.String("data", string(p.Body)))
				re.Message = "切换房间失败"
			} else {
				re.Message = s.Message()
			}
			p.Body, _ = json.Marshal(re)
			return nil
		}

		if err := b.ChangeRoom(r.RoomId, ch); err != nil {
			log.Error("change room", zap.Error(err), zap.String("data", string(p.Body)))
			re := newConnect()
			re.Message = "切换房间失败"
			p.Body, _ = json.Marshal(re)
			return nil
		}

		reply.Connect.Key = ch.Key
		p.Body, _ = json.Marshal(reply.Connect)

		if reply.TopMessage != nil {
			ch.protoRing.SetAdv()
			ch.Signal()

			p1, err := ch.protoRing.Set()
			if err != nil {
				log.Error("proto ping set top message for change room")
				return nil
			}

			p1.Op = cometpb.OpRaw
			p1.Body = reply.TopMessage
		}
		if reply.BulletinMessage != nil {
			ch.protoRing.SetAdv()
			ch.Signal()

			p1, err := ch.protoRing.Set()
			if err != nil {
				log.Error("proto ping set bulletin message for change room")
				return nil
			}

			p1.Op = cometpb.OpRaw
			p1.Body = reply.BulletinMessage
		}

		if reply.IsConnectSuccessReply {
			if _, e := s.ConnectSuccessReply(ctx, ch.Room.ID, reply.User, reply.Connect); e != nil {
				log.Error("connect success reply", zap.Error(e), zap.Int32("rid", ch.Room.ID), zap.Any("user", reply.User))
			}
		}

	default:
		// TODO error
		p.Body = nil
	}
	return nil
}

func newConnect() *logicpb.Connect {
	return &logicpb.Connect{
		Permission:        new(logicpb.Permission),
		PermissionMessage: new(logicpb.PermissionMessage),
	}
}
