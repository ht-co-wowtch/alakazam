package dao

import "database/sql"

type store struct {
	*sql.DB
}

