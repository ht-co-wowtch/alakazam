package client

import (
	"encoding/json"
	"fmt"
	"net/http"

	"gitlab.com/ht-co/cpw/micro/log"
)

type LiveChatInfo struct {
	Id        int     `json:"id"`
	SiteId    int     `json:"site_id"`
	ChatId    int     `json:"chat_id"`
	MemberId  int     `json:"member_id"`
	MemberUid string  `json:"member_uid"`
	IsLive    bool    `json:"is_live"`
	Charge    float32 `json:"charge"`
	RoomId    string  `json:"room_id"`
}

// 付費房收費標準
// LiveChatCharge
func (c *Client) GetLiveChatInfo(roomID int32) (LiveChatInfo, error) {
	path := fmt.Sprintf("/live/chat/%d", roomID)
	resp, _ := c.c.Get(path, nil, nil)

	var lci LiveChatInfo
	if err := json.NewDecoder(resp.Body).Decode(&lci); err != nil {
		log.Errorf("get LiveChatCharge error:, %o", err)
		return LiveChatInfo{}, err
	}

	return lci, nil
}

// 建立付費房付費訂單
// CreateLiveChatPaidOrder
func (c *Client) CreateLiveChatPaidOrder(siteId int, wmUID string, lcID int, orderId string, amount float32) (bool, error) {
	resp, err := c.c.PostJson(
		"/live/order",
		nil,
		struct {
			SiteId         int     `json:"site_id"`
			WithdrawMember string  `json:"withdraw_member"`
			LiveChatId     int     `json:"live_chat_id"`
			OrderId        string  `json:"order_id"`
			Amount         float32 `json:"amount"`
		}{
			SiteId:         siteId,
			WithdrawMember: wmUID,
			LiveChatId:     lcID,
			OrderId:        orderId,
			Amount:         amount,
		},
		nil)

	if resp.StatusCode != http.StatusOK {
		return false, err
	}

	return true, nil
}

type PaidDiamondUser struct {
	Uid  string `json:"uid"`
	Type string `json:"type"`
}

type PaidDiamondOrder struct {
	Id     string  `json:"id"`
	Amount float32 `json:"amount"`
}

type PaidDiamondTXTOrder struct {
	From   PaidDiamondUser    `json:"from"`
	To     PaidDiamondUser    `json:"to"`
	Orders []PaidDiamondOrder `json:"orders"`
}

type TxUserResp struct {
	Uid     string  `json:"uid"`
	Balance float32 `json:"balance"`
	Lien    float32 `json:"lien"`
	Diamond float32 `json:"diamond"`
}

type TxtResp struct {
	From TxUserResp `json:"from"`
	To   TxUserResp `json:"to"`
}

// 轉帳異動鑽石餘額
// PaidDiamond
func (c *Client) PaidDiamond(orders PaidDiamondTXTOrder) (TxtResp, error) {
	resp, err := c.c.PostJson("/txt/diamond", nil, orders, nil)
	if err != nil {
		return TxtResp{}, err
	}

	if resp.StatusCode != http.StatusOK {
		// TODO
		e := WalletError{}
		defer resp.Body.Close()
		if err := json.NewDecoder(resp.Body).Decode(&e); err != nil {
			return TxtResp{}, err
		}

		return TxtResp{}, e
	}

	var tr TxtResp
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return TxtResp{}, err
	}

	return tr, nil
}

type WalletError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e WalletError) Error() string {
	return e.Message
}
