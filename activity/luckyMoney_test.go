package activity

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"testing"
)

func TestGiveMoney(t *testing.T) {
	mockMoneyApi := new(mockMoneyApi)

	amount := float32(2)
	count := 2

	uuid := mock.MatchedBy(func(id string) bool {
		_, err := uuid.Parse(id)
		return err == nil
	})

	mockMoneyApi.On("NewOlder", uuid, float32(4), "1").
		Return(nil)

	err := getMock(mockMoneyApi).Give(&GiveMoney{
		Amount:  amount,
		Count:   count,
		Message: "test",
		Type:    Money,
		Token:   "1",
	})

	mockMoneyApi.AssertExpectations(t)

	assert.Nil(t, err)
}

func TestGiveMoneyByBalanceError(t *testing.T) {
	mockMoneyApi := new(mockMoneyApi)

	mockMoneyApi.On("NewOlder", mock.Anything, mock.Anything, mock.Anything).
		Return(errors.BalanceError)

	err := getMock(mockMoneyApi).Give(&GiveMoney{
		Type: Money,
	})

	mockMoneyApi.AssertExpectations(t)

	assert.Equal(t, errors.BalanceError, err)
}

type mockMoneyApi struct {
	mock.Mock
}

func getMock(m moneyApi) *LuckyMoney {
	return &LuckyMoney{
		money: m,
	}
}

func (m *mockMoneyApi) NewOlder(id string, total float32, token string) error {
	args := m.Called(id, total, token)
	return args.Error(0)
}
