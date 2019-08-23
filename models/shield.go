package models

import "database/sql"

type Shield struct {
	Id      int `xorm:"pk autoincr"`
	Context string
}

func (s *Shield) TableName() string {
	return "shields"
}

func (s *Store) GetShield(id int) (Shield, error) {
	shield := Shield{}
	ok, err := s.d.Where("`id` = ?", id).Get(&shield)
	if err != nil {
		return Shield{}, err
	}
	if !ok {
		return Shield{}, sql.ErrNoRows
	}
	return shield, nil
}

func (s *Store) GetShieldContext(context string) (Shield, error) {
	shield := Shield{}
	ok, err := s.d.Where("`context` = ?", context).Get(&shield)
	if err != nil {
		return Shield{}, err
	}
	if !ok {
		return Shield{}, sql.ErrNoRows
	}
	return shield, nil
}

func (s *Store) GetAllShield() ([]Shield, error) {
	shields := make([]Shield, 0)
	err := s.d.Find(&shields)
	return shields, err
}

func (s *Store) CreateShield(context string) (*Shield, error) {
	shield := &Shield{
		Context: context,
	}
	aff, err := s.d.InsertOne(shield)
	if err != nil {
		return nil, err
	}
	if aff != 1 {
		return nil, ErrInsertFailure
	}
	return shield, nil
}

func (s *Store) UpdateShield(id int, context string) error {
	shield := Shield{
		Context: context,
	}
	_, err := s.d.Where("`id` = ?", id).Update(&shield)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) DeleteShield(id int) error {
	aff, err := s.d.Where("`id` = ?", id).Delete(&Shield{})
	if err != nil {
		return err
	}
	if aff != 1 {
		return ErrDeleteFailure
	}
	return nil
}
