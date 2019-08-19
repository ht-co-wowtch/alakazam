package message

import (
	"context"
	"encoding/json"
	"gitlab.com/jetfueltw/cpw/alakazam/app/logic/pb"
	seqpb "gitlab.com/jetfueltw/cpw/alakazam/app/seq/api/pb"
	"time"
)

type Message struct {
	Id      int64  `json:"id"`
	Uid     string `json:"uid"`
	Name    string `json:"name"`
	Avatar  string `json:"avatar"`
	Message string `json:"message"`
	Time    string `json:"time"`
}

type Messages struct {
	Rooms   []string
	Rids    []int64
	Mid     int64
	Uid     string
	Name    string
	Message string
}

func (p *Producer) Send(message Messages) (int64, error) {
	seq, err := p.seq.Id(context.Background(), &seqpb.Arg{
		Code: message.Rids[0], Count: 1,
	})
	if err != nil {
		return 0, err
	}

	bm, err := json.Marshal(Message{
		Id:      seq.Id,
		Uid:     message.Uid,
		Name:    message.Name,
		Message: message.Message,
		Time:    time.Now().Format(time.RFC3339),
	})
	if err != nil {
		return 0, err
	}
	pushMsg := &pb.PushMsg{
		Type: pb.PushMsg_ROOM,
		Room: message.Rooms,
		Mid:  message.Mid,
		Rids: message.Rids,
		Msg:  bm,
	}
	if err := p.send(pushMsg); err != nil {
		return 0, err
	}
	return seq.Id, nil
}

type RedEnvelopeMessage struct {
	Messages
	RedEnvelopeId string
	Token         string
	Expired       int64
}

type money struct {
	Message
	RedEnvelope
}

type RedEnvelope struct {
	Id      string `json:"id"`
	Token   string `json:"token"`
	Expired int64  `json:"expired"`
}

func (p *Producer) SendRedEnvelope(message RedEnvelopeMessage) error {
	bm, err := json.Marshal(money{
		Message: Message{
			Uid:     message.Uid,
			Name:    message.Name,
			Message: message.Message,
			Time:    time.Now().Format(time.RFC3339),
		},
		RedEnvelope: RedEnvelope{
			Id:      message.RedEnvelopeId,
			Token:   message.Token,
			Expired: message.Expired,
		},
	})
	if err != nil {
		return err
	}
	pushMsg := &pb.PushMsg{
		Type: pb.PushMsg_MONEY,
		Room: message.Rooms,
		Mid:  message.Mid,
		Rids: message.Rids,
		Msg:  bm,
	}
	if err := p.send(pushMsg); err != nil {
		return err
	}
	return nil
}

type AdminMessage struct {
	Rooms   []string
	Rids    []int64
	Message string
	IsTop   bool
}

// 所有房間推送
// TODO 需實作訊息是否頂置
func (p *Producer) SendForAdmin(message AdminMessage) (int64, error) {
	msg, err := json.Marshal(Message{
		Name:    "管理员",
		Message: message.Message,
		Time:    time.Now().Format(time.RFC3339),
	})
	if err != nil {
		return 0, err
	}

	var t pb.PushMsg_Type
	if message.IsTop {
		t = pb.PushMsg_TOP
	} else {
		t = pb.PushMsg_ROOM
	}

	pushMsg := &pb.PushMsg{
		Type: t,
		Room: message.Rooms,
		Rids: message.Rids,
		Msg:  msg,
	}
	err = p.send(pushMsg)
	if err != nil {
		return 0, err
	}
	return 0, nil
}
