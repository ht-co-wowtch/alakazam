package models

import (
	"errors"
	"sync"
)

type Seq struct {
	Id     int64 `xorm:"pk"`
	RoomId int64
	Max    int64
	Cur    int64 `xorm:"-"`
	Batch  int64
	L      sync.Mutex `xorm:"-"`
}

type ISeq interface {
	Sync(seq *Seq) (bool, error)
	LoadSeq() ([]Seq, error)
	Create(code, batch int64) error
}

func (r *Seq) TableName() string {
	return "seqs"
}

func (d *Store) Sync(seq *Seq) (bool, error) {
	aff, err := d.d.Where("room_id = ?", seq.RoomId).
		Cols("max").
		Update(seq)
	return aff >= 1, err
}

func (d *Store) LoadSeq() ([]Seq, error) {
	seq := make([]Seq, 0)
	err := d.d.Table(&Seq{}).Find(&seq)
	return seq, err
}

func (d *Store) Create(code, batch int64) error {
	s := Seq{
		RoomId: code,
		Batch:  batch,
	}
	aff, err := d.d.InsertOne(&s)
	if err != nil {
		return err
	}
	if aff != 1 {
		return errors.New("insert failure")
	}
	return nil
}
