package client

type Money struct {
	Dml     int `json:"dml"`
	Deposit int `json:"deposit"`
}

// TODO 待實作
func (c *Client) GetDepositAndDml(day int, uid, token string) (money Money, err error) {
	return money, err
}

type Older struct {
	// 訂單編號 char(32)
	OrderId string `json:"order_id"`

	// 金額
	Amount float64 `json:"amount"`
}

type olderReply struct {
	Balance float64 `json:"balance"`
}

// TODO 待實作
func (c *Client) NewOlder(older Older, uid, token string) (float64, error) {
	return 0, nil
}
