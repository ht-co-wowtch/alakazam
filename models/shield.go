package models

type Shield struct {
	Id      int `xorm:"pk autoincr"`
	Context string
}

func (s *Shield) TableName() string {
	return "shields"
}

func (s *Store) CreateShield(context string) (int, error) {
	shield := Shield{
		Context: context,
	}
	aff, err := s.d.InsertOne(&shield)
	if err != nil {
		return 0, err
	}
	if aff != 1 {
		return 0, ErrInsertFailure
	}
	return shield.Id, nil
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
