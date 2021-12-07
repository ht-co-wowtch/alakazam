package api

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"gitlab.com/jetfueltw/cpw/alakazam/app/seq/api/pb"
	"gitlab.com/jetfueltw/cpw/alakazam/app/seq/conf"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	rpc "gitlab.com/ht-co/cpw/micro/grpc"
	"gitlab.com/ht-co/cpw/micro/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"time"
)

func NewServer(ctx context.Context, c *conf.Config) (*grpc.Server, error) {
	srv := rpc.New(c.RPCServer)
	bs := make(map[int64]*models.Seq, 0)
	db := models.NewStore(c.DB)
	rpcS := rpcServer{
		bs: bs,
		db: db,
	}
	if err := rpcS.Load(); err != nil {
		return nil, err
	}
	go func() {
		t := time.NewTicker(time.Minute * 10)

		for {
			select {
			case <-t.C:
				if err := rpcS.Load(); err != nil {
					log.Error("load seq", zap.Error(err))
				} else {
					log.Info("load seq")
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	pb.RegisterSeqServer(srv, &rpcS)
	return srv, nil
}

type rpcServer struct {
	bs map[int64]*models.Seq
	db models.ISeq
}

func (s *rpcServer) Ping(context.Context, *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}

func (s *rpcServer) Id(ctx context.Context, arg *pb.SeqReq) (*pb.SeqResp, error) {
	arg.Count = 1
	return s.Ids(ctx, arg)
}

func (s *rpcServer) Ids(ctx context.Context, arg *pb.SeqReq) (*pb.SeqResp, error) {
	b, ok := s.bs[arg.Id]
	if !ok {
		return nil, errors.New("not seq code")
	}
	var seq int64
	b.Mu.Lock()
	defer b.Mu.Unlock()
	if seq = b.Cur + arg.Count; seq >= b.Max {
		b.Max += b.Batch
		ok, err := s.db.SyncSeq(b)
		if err != nil {
			log.Error("grpc Ids", zap.Error(err), zap.Int64("id", arg.Id), zap.Int64("count", arg.Count))
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("business model sync id: [%d] seq: [%d]", arg.Id, b.Cur)
		}
	}
	b.Cur = seq
	return &pb.SeqResp{Id: b.Cur}, nil
}

func (s *rpcServer) Create(ctx context.Context, info *pb.SeqCreateReq) (*pb.SeqCreateResp, error) {
	id, err := s.db.CreateSeq(info.Batch)
	if err != nil {
		log.Error("grpc create", zap.Error(err))
		return &pb.SeqCreateResp{}, err
	}
	return &pb.SeqCreateResp{Id: int64(id)}, nil
}

func (s *rpcServer) Load() error {
	seqs, err := s.db.LoadSeq()
	if err != nil {
		return err
	}
	for _, v := range seqs {
		v.Mu.Lock()
		v.Cur = v.Max
		s.bs[int64(v.Id)] = &v
		v.Mu.Unlock()
	}
	return nil
}
