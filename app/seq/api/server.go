package api

import (
	"context"
	"fmt"
	"gitlab.com/jetfueltw/cpw/alakazam/app/seq/api/pb"
	"gitlab.com/jetfueltw/cpw/alakazam/app/seq/conf"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	rpc "gitlab.com/jetfueltw/cpw/micro/grpc"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func NewServer(c *conf.Config) (*grpc.Server, error) {
	srv := rpc.New(c.RPCServer)
	bs := make(map[int64]*models.Seq, 0)
	db := models.NewStore(c.DB)
	seqs, err := db.LoadSeq()
	if err != nil {
		return nil, err
	}
	for _, v := range seqs {
		v.Cur = v.Max
		bs[int64(v.Id)] = &v
	}
	pb.RegisterSeqServer(srv, &rpcServer{
		bs: bs,
		db: db,
	})
	return srv, nil
}

type rpcServer struct {
	bs map[int64]*models.Seq
	db models.ISeq
}

func (s *rpcServer) Ping(context.Context, *pb.Empty) (*pb.Empty, error) {
	return &pb.Empty{}, nil
}

func (s *rpcServer) Id(ctx context.Context, arg *pb.Arg) (*pb.SeqId, error) {
	arg.Count = 1
	return s.Ids(ctx, arg)
}

func (s *rpcServer) Ids(ctx context.Context, arg *pb.Arg) (*pb.SeqId, error) {
	b := s.bs[arg.Code]
	var seq int64
	b.L.Lock()
	defer b.L.Unlock()
	if seq = b.Cur + arg.Count; seq >= b.Max {
		b.Max += b.Batch
		ok, err := s.db.SyncSeq(b)
		if err != nil {
			log.Error("grpc Ids", zap.Error(err), zap.Int64("code", arg.Code), zap.Int64("count", arg.Count))
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("business model sync id: [%d] seq: [%d]", arg.Code, b.Cur)
		}
	}
	b.Cur = seq
	return &pb.SeqId{Id: b.Cur}, nil
}

func (s *rpcServer) Create(ctx context.Context, info *pb.Info) (*pb.Empty, error) {
	if err := s.db.CreateSeq(info.Code, info.Batch); err != nil {
		log.Error("grpc create", zap.Error(err), zap.Int64("code", info.Code))
		return &pb.Empty{}, err
	}
	return &pb.Empty{}, nil
}
