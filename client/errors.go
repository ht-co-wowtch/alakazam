package client

import (
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"net/http"
)

var (
	InsufficientBalanceError = errors.New(http.StatusPaymentRequired, 15024020, "Insufficient balance")
)
