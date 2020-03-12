package message

import "encoding/json"

type System struct {
	Id        int64                  `json:"id"`
	Type      string                 `json:"type"`
	Name      string                 `json:"name"`
	Message   string                 `json:"message"`
	Time      string                 `json:"time"`
	Timestamp int64                  `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

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

type Bets struct {
	Id        int64  `json:"id"`
	Uid       string `json:"uid"`
	Type      string `json:"type"`
	Name      string `json:"name"`
	Avatar    string `json:"avatar"`
	Time      string `json:"time"`
	Timestamp int64  `json:"timestamp"`

	GameId       int   `json:"game_id"`
	PeriodNumber int   `json:"period_number"`
	Items        []Bet `json:"bets"`
	Count        int   `json:"count"`
	TotalAmount  int   `json:"total_amount"`
}

type Bet struct {
	Name       string   `json:"name"`
	OddsCode   string   `json:"odds_code"`
	Items      []string `json:"items"`
	TransItems []string `json:"trans_items"`
	Amount     int      `json:"amount"`
}

type Gift struct {
	Id          int64  `json:"id"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	Message     string `json:"message"`
	Animation   string `json:"animation"`
	AnimationId int    `json:"animation_id"`
	Time        string `json:"time"`
	Timestamp   int64  `json:"timestamp"`
}

type Reward struct {
	Id        int64  `json:"id"`
	Type      string `json:"type"`
	Name      string `json:"name"`
	Message   string `json:"message"`
	Time      string `json:"time"`
	Timestamp int64  `json:"timestamp"`
}
