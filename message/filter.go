package message

import (
	"database/sql"
	"gitlab.com/ht-co/wowtch/live/alakazam/errors"
	"gitlab.com/ht-co/wowtch/live/alakazam/models"
	shield "gitlab.com/ht-co/wowtch/live/alakazam/pkg/filter"
)

type Filter interface {
	Create(context string) (int, error)
	Update(id int, context string) error
	Delete(id int) error
}

type filter struct {
	db     *models.Store
	filter *shield.Filter
}

func NewFilter(db *models.Store) Filter {
	return newFilter(db)
}

func newFilter(db *models.Store) *filter {
	return &filter{
		db:     db,
		filter: shield.New(),
	}
}

func (s *filter) load() error {
	shields, err := s.db.GetAllShield()
	if err != nil {
		return err
	}
	str := make([]string, len(shields))
	for _, v := range shields {
		str = append(str, v.Context)
	}
	s.filter.Adds(str)
	return nil
}

func (s *filter) Create(context string) (int, error) {
	str, err := s.db.GetShieldContext(context)
	if err == sql.ErrNoRows {
		st, err := s.db.CreateShield(context)
		if err != nil {
			return 0, err
		}
		s.filter.Add(context)
		return st.Id, nil
	} else if err != nil {
		return 0, err
	} else if str.Id != 0 {
		return 0, errors.ErrExist
	}
	return str.Id, nil
}

func (s *filter) Update(id int, context string) error {
	_, err := s.db.GetShield(id)
	if err == sql.ErrNoRows {
		return errors.ErrNoRows
	}
	if err != nil {
		return err
	}
	err = s.db.UpdateShield(id, context)
	if err == nil {
		s.filter.Add(context)
	}
	return err
}

func (s *filter) Delete(id int) error {
	str, err := s.db.GetShield(id)
	if err == sql.ErrNoRows {
		return errors.ErrNoRows
	}
	if err != nil {
		return err
	}
	if err = s.db.DeleteShield(id); err != nil {
		return err
	}
	s.filter.Delete(str.Context)
	return nil
}
