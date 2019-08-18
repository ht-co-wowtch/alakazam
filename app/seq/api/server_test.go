package api

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/jetfueltw/cpw/alakazam/app/seq/api/pb"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"testing"
)

var (
	fakeSeq = []models.Seq{
		models.Seq{
			Id:    1,
			Cur:   0,
			Max:   10,
			Batch: 2,
		},
	}
)

func TestGenerateId(t *testing.T) {
	seq, _ := mockSeq()
	id, err := seq.Id(context.TODO(), &pb.Arg{Code: 1})

	assert.Nil(t, err)
	assert.Equal(t, int64(1), id.Id)
}

func TestGenerateIdByIncrement(t *testing.T) {
	seq, _ := mockSeq()

	i, err := seq.Id(context.TODO(), &pb.Arg{Code: 1})
	ii, erri := seq.Id(context.TODO(), &pb.Arg{Code: 1})

	assert.Nil(t, err)
	assert.Nil(t, erri)
	assert.Equal(t, int64(1), i.Id)
	assert.Equal(t, int64(2), ii.Id)
}

func TestGenerateIds(t *testing.T) {
	seq, _ := mockSeq()

	id, err := seq.Ids(context.TODO(), &pb.Arg{Code: 1, Count: 2})

	assert.Nil(t, err)
	assert.Equal(t, int64(2), id.Id)
}

func TestSeqSync(t *testing.T) {
	seq, dao := mockSeq()

	dao.On("Sync", mock.MatchedBy(func(m *models.Seq) bool {
		return int(m.Max) == (10 + m.Batch)
	})).Once().Return(true, nil)

	_, err := seq.Ids(context.TODO(), &pb.Arg{Code: 1, Count: 10})

	assert.Nil(t, err)
}

func mockSeq() (*rpcServer, *mockDao) {
	m := &mockDao{}
	bs := make(map[int64]*models.Seq, 0)
	for _, v := range fakeSeq {
		bs[int64(v.Id)] = &v
	}
	return &rpcServer{
		db: m,
		bs: bs,
	}, m
}

type mockDao struct {
	mock.Mock
}

func (m mockDao) SyncSeq(business *models.Seq) (bool, error) {
	arg := m.Called(business)
	return arg.Bool(0), arg.Error(1)
}

func (m mockDao) LoadSeq() ([]models.Seq, error) {
	return fakeSeq, nil
}

func (m mockDao) CreateSeq(code, batch int64) error {
	arg := m.Called(code, batch)
	return arg.Error(0)
}
