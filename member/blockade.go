package member

func (m *Member) SetBlockade(uid string, rid int, set bool) error {
	member, err := m.GetByRoom(uid, rid)
	if err != nil {
		return err
	}

	if member.IsBlockade == set {
		return nil
	}

	member.IsBlockade = set

	if err := m.db.SetPermission(*member); err != nil {
		return err
	}

	return m.c.set(member)
}
