package message

import "gitlab.com/jetfueltw/cpw/alakazam/app/logic/pb"

type Message struct {
	Id      int64           `json:"id"`
	Uid     string          `json:"uid"`
	Type    pb.PushMsg_Type `json:"type"`
	Name    string          `json:"name"`
	Avatar  string          `json:"avatar"`
	Message string          `json:"message"`
	Time    string          `json:"time"`
}

type Money struct {
	Message
	RedEnvelope
}

type RedEnvelope struct {
	Id      string `json:"red_envelope_id"`
	Token   string `json:"token"`
	Expired int64  `json:"expired"`
}
