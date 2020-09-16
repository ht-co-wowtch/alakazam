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
	CreateUser(member *Member) (bool, error)
	UpdateUser(member *Member) (bool, error)
	SetUserBlockade(uid string, is bool) (bool, error)
	SetUserBanned(uid string, is bool) (bool, error)
	GetMembers(ids []int64) ([]Member, error)
	GetMembersByUid(uid []string) ([]Member, error)
	SetBannedLog(mid int64, sec time.Duration, isSystem bool) (bool, error)
	GetTodaySystemBannedLog(mid int64) ([]BannedLog, error)
	RoomPermission(id int64, rid int) (Permission, error)
	SetRoomPermission(member Member) error
}

type Store struct {
	d *xorm.EngineGroup
}

func NewStore(c *database.Conf) *Store {
	return &Store{
		d: NewORM(c),
	}
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
