package message

import (
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
)

type Shield interface {
	Create(context string) (int, error)
	Update(id int, context string) error
	Delete(id int) error
}

type shield struct {
	db *models.Store
}

func NewShield(db *models.Store) Shield {
	return &shield{
		db: db,
	}
}

func (s *shield) Create(context string) (int, error) {
	return s.db.CreateShield(context)
}

func (s *shield) Update(id int, context string) error {
	return s.db.UpdateShield(id, context)
}

func (s *shield) Delete(id int) error {
	err := s.db.DeleteShield(id)
	if err == models.ErrDeleteFailure {
		return errors.ErrNoRows
	}
	return err
}
