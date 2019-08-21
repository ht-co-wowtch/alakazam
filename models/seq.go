package models

import (
	"errors"
	"sync"
)

type Seq struct {
	Id    int `xorm:"pk"`
	Max   int64
	Cur   int64 `xorm:"-"`
	Batch int64
	L     sync.Mutex `xorm:"-"`
}

type ISeq interface {
	SyncSeq(seq *Seq) (bool, error)
	LoadSeq() ([]Seq, error)
	CreateSeq(batch int64) (int, error)
}

func (r *Seq) TableName() string {
	return "seqs"
}

func (d *Store) SyncSeq(seq *Seq) (bool, error) {
	aff, err := d.d.Where("id = ?", seq.Id).
		Cols("max").
		Update(seq)
	return aff >= 1, err
}

func (d *Store) LoadSeq() ([]Seq, error) {
	seq := make([]Seq, 0)
	err := d.d.Table(&Seq{}).Find(&seq)
	return seq, err
}

func (d *Store) CreateSeq(batch int64) (int, error) {
	s := Seq{
		Batch: batch,
	}
	aff, err := d.d.Master().InsertOne(&s)
	if err != nil {
		return 0, err
	}
	if aff != 1 {
		return 0, errors.New("insert failure")
	}
	return s.Id, nil
}
