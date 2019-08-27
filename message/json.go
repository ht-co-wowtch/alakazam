package message

type Message struct {
	Id      int64  `json:"id"`
	Uid     string `json:"uid"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Avatar  string `json:"avatar"`
	Message string `json:"message"`
	Time    string `json:"time"`
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
	Id      int64  `json:"id"`
	Uid     string `json:"uid"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Avatar  string `json:"avatar"`
	Message string `json:"message"`
	Time    string `json:"time"`
}

type historyRedEnvelopeMessage struct {
	historyMessage
	RedEnvelope historyRedEnvelope `json:"red_envelope"`
}

type historyRedEnvelope struct {
	Id      string `json:"id"`
	Token   string `json:"token"`
	Expired string `json:"expired"`
}
