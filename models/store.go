package models

import (
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"gitlab.com/jetfueltw/cpw/micro/database"
	"time"
)

var tables []interface{}

var (
	ErrInsertFailure = errors.New("insert failure")
	ErrUpdateFailure = errors.New("update failure")
	ErrDeleteFailure = errors.New("delete failure")
)

func init() {
	tables = append(tables,
		new(Member),
		new(Room),
		new(RedEnvelopeMessage),
		new(Message),
		new(RoomMessage),
		new(RoomTopMessage),
	)
}

type Chat interface {
	Find(uid string) (*Member, error)
	Permission(id int64, rid int) (Permission, error)
	SetPermission(member Member) error
	CreateUser(member *Member) (bool, error)
	UpdateUser(member *Member) (bool, error)
	GetMembers(ids []int64) ([]Member, error)
	GetMembersByUid(uid []string) ([]Member, error)
	SetBlockade(uid string) (int64, error)
	DeleteBanned(uid string) (int64, error)
	SetBannedLog(mid int64, sec time.Duration, isSystem bool) (bool, error)
	GetTodaySystemBannedLog(mid int64) ([]BannedLog, error)
}

type Store struct {
	d *xorm.EngineGroup
}

func Table() []interface{} {
	return tables
}

func NewStore(c *database.Conf) *Store {
	return &Store{
		d: NewORM(c),
	}
}

func NewChat(c *database.Conf) Chat {
	return NewStore(c)
}

func NewORM(c *database.Conf) *xorm.EngineGroup {
	d, err := database.NewORM(c)
	if err != nil {
		panic(err)
	}

	l, err := time.LoadLocation(c.Master.Local)
	if err != nil {
		panic(err)
	}

	d.Master().SetTZDatabase(l)
	d.Slave().SetTZDatabase(l)
	return d
}
