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

type Money struct {
	Message
	RedEnvelope RedEnvelope `json:"red_envelope"`
}

type RedEnvelope struct {
	Id      string `json:"id"`
	Token   string `json:"token"`
	Expired string `json:"expired"`
}

type historyMessage struct {
	Id        int64  `json:"id"`
	Uid       string `json:"uid"`
	Type      string `json:"type"`
	Name      string `json:"name"`
	Avatar    string `json:"avatar"`
	Message   string `json:"message"`
	Time      string `json:"time"`
	Timestamp int64  `json:"timestamp"`
}

func (m historyMessage) MarshalBinary() (data []byte, err error) {
	return json.Marshal(m)
}

func (m historyMessage) Score() float64 {
	return float64(m.Timestamp)
}

type historyRedEnvelopeMessage struct {
	historyMessage
	RedEnvelope historyRedEnvelope `json:"red_envelope"`
}

func (m historyRedEnvelopeMessage) Score() float64 {
	return float64(m.historyMessage.Timestamp)
}

type historyRedEnvelope struct {
	Id      string `json:"id"`
	Token   string `json:"token"`
	Expired string `json:"expired"`
}

func (m historyRedEnvelope) MarshalBinary() (data []byte, err error) {
	return json.Marshal(m)
}
