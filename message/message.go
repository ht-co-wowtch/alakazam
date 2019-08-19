package message

import (
	"context"
	"encoding/json"
	"gitlab.com/jetfueltw/cpw/alakazam/app/logic/pb"
	seqpb "gitlab.com/jetfueltw/cpw/alakazam/app/seq/api/pb"
	"time"
)

const (
	RootMid  = 1
	RootUid  = "root"
	RootName = "管理员"
)

type Messages struct {
	Rooms   []string
	Rids    []int64
	Mid     int64
	Uid     string
	Name    string
	Message string
	IsTop   bool
}

func (p *Producer) toPb(msg Messages) (*pb.PushMsg, error) {
	seq, err := p.seq.Id(context.Background(), &seqpb.Arg{
		Code: msg.Rids[0], Count: 1,
	})
	if err != nil {
		return nil, err
	}
	now := time.Now()
	bm, err := json.Marshal(Message{
		Id:      seq.Id,
		Type:    pb.PushMsg_ROOM,
		Uid:     msg.Uid,
		Name:    msg.Name,
		Message: msg.Message,
		Time:    now.Format(time.RFC3339),
	})
	if err != nil {
		return nil, err
	}
	return &pb.PushMsg{
		Seq:     seq.Id,
		Type:    pb.PushMsg_ROOM,
		Room:    msg.Rooms,
		Mid:     msg.Mid,
		Rids:    msg.Rids,
		Msg:     bm,
		Message: msg.Message,
		SendAt:  now.Unix(),
	}, nil
}

func (p *Producer) Send(msg Messages) (int64, error) {
	pushMsg, err := p.toPb(msg)
	if err != nil {
		return 0, err
	}
	if err := p.send(pushMsg); err != nil {
		return 0, err
	}
	return pushMsg.Seq, nil
}

type AdminMessage struct {
	Rooms   []string
	Rids    []int64
	Message string
	IsTop   bool
}

// 所有房間推送
// TODO 需實作訊息是否頂置
func (p *Producer) SendForAdmin(msg AdminMessage) (int64, error) {
	pushMsg, err := p.toPb(Messages{
		Rooms:   msg.Rooms,
		Rids:    msg.Rids,
		Mid:     RootMid,
		Uid:     RootUid,
		Name:    RootName,
		Message: msg.Message,
	})
	if err != nil {
		return 0, err
	}
	if msg.IsTop {
		pushMsg.Type = pb.PushMsg_TOP
	}
	if err := p.send(pushMsg); err != nil {
		return 0, err
	}
	return pushMsg.Seq, nil
}

type RedEnvelopeMessage struct {
	Messages
	RedEnvelopeId string
	Token         string
	Expired       int64
}

func (p *Producer) toRedEnvelopePb(msg RedEnvelopeMessage) (*pb.PushMsg, error) {
	seq, err := p.seq.Id(context.Background(), &seqpb.Arg{
		Code: msg.Rids[0], Count: 1,
	})
	if err != nil {
		return nil, err
	}
	now := time.Now()
	bm, err := json.Marshal(money{
		Message: Message{
			Id:      seq.Id,
			Type:    pb.PushMsg_MONEY,
			Uid:     msg.Uid,
			Name:    msg.Name,
			Message: msg.Message,
			Time:    now.Format(time.RFC3339),
		},
		RedEnvelope: RedEnvelope{
			Id:      msg.RedEnvelopeId,
			Token:   msg.Token,
			Expired: msg.Expired,
		},
	})
	if err != nil {
		return nil, err
	}
	return &pb.PushMsg{
		Seq:     seq.Id,
		Type:    pb.PushMsg_MONEY,
		Room:    msg.Rooms,
		Mid:     msg.Mid,
		Rids:    msg.Rids,
		Msg:     bm,
		SendAt:  now.Unix(),
		Message: msg.Message,
	}, nil
}

func (p *Producer) SendRedEnvelope(msg RedEnvelopeMessage) (int64, error) {
	pushMsg, err := p.toRedEnvelopePb(msg)
	if err != nil {
		return 0, err
	}
	if err := p.send(pushMsg); err != nil {
		return 0, err
	}
	return pushMsg.Seq, nil
}

type AdminRedEnvelopeMessage struct {
	AdminMessage
	RedEnvelopeId string
	Token         string
	Expired       int64
}

func (p *Producer) SendRedEnvelopeForAdmin(msg AdminRedEnvelopeMessage) (int64, error) {
	pushMsg, err := p.toRedEnvelopePb(RedEnvelopeMessage{
		Messages: Messages{
			Rooms:   msg.Rooms,
			Rids:    msg.Rids,
			Mid:     RootMid,
			Uid:     RootUid,
			Name:    RootName,
			Message: msg.Message,
		},
		RedEnvelopeId: msg.RedEnvelopeId,
		Token:         msg.Token,
		Expired:       msg.Expired,
	})
	if err != nil {
		return 0, err
	}
	if err := p.send(pushMsg); err != nil {
		return 0, err
	}
	return pushMsg.Seq, nil
}
