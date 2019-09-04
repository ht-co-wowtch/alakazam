package message

import "encoding/json"

type Message struct {
	Id        int64  `json:"id"`
	Uid       string `json:"uid"`
	Type      string `json:"type"`
	Name      string `json:"name"`
	Avatar    string `json:"avatar"`
	Message   string `json:"message"`
	Time      string `json:"time"`
	Timestamp int64  `json:"timestamp"`
}

func (m Message) MarshalBinary() (data []byte, err error) {
	return json.Marshal(m)
}

func (m Message) Score() float64 {
	return float64(m.Timestamp)
}

type RedEnvelopeMessage struct {
	Message
	RedEnvelope RedEnvelope `json:"red_envelope"`
}

func (m RedEnvelopeMessage) Score() float64 {
	return float64(m.Message.Timestamp)
}

type RedEnvelope struct {
	Id      string `json:"id"`
	Token   string `json:"token"`
	Expired string `json:"expired"`
}

func (m RedEnvelope) MarshalBinary() (data []byte, err error) {
	return json.Marshal(m)
}
