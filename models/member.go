package models

import (
	"database/sql"
	"time"

	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
)

const (
	// 封鎖
	Blockade = 0

	// 聊天
	MessageStatus = 1

	// 搶紅包
	redEnvelope = 2

	PlayStatus = MessageStatus + redEnvelope
)

const (
	// 訪客
	Guest = 0

	// 營銷
	Market = 1

	// 玩家
	Player = 2

	// 直播主
	STREAMER = 3

	// 房管
	MANAGE = 4
)

type Member struct {
	Id         int64      `xorm:"pk autoincr"`
	Uid        string     `json:"uid"`
	Name       string     `json:"name"`
	Type       int        `json:"type"`
	IsMessage  bool       `json:"is_message"`
	IsBlockade bool       `json:"is_blockade"`
	Permission Permission `json:"-" xorm:"-"`
	Gender     int32      `json:"gender"`
	CreateAt   time.Time  `json:"-"`
}

func (r *Member) TableName() string {
	return "members"
}

func (r Member) Banned() bool {
	if r.IsMessage {
		return r.Permission.IsBanned
	}
	return true
}

func (r Member) Blockade() bool {
	if !r.IsBlockade {
		return r.Permission.IsBlockade
	}
	return true
}

type Permission struct {
	RoomId     int64
	MemberId   int64
	IsBanned   bool
	IsBlockade bool
	IsManage   bool
}

func (r *Permission) TableName() string {
	return "room_user_permissions"
}

// 新增會員
func (s *Store) CreateUser(member *Member) (bool, error) {
	member.CreateAt = time.Now()
	member.IsMessage = true
	aff, err := s.d.InsertOne(member)
	return aff == 1, err
}

func (s *Store) UpdateUser(member *Member) (bool, error) {
	aff, err := s.d.Cols("name", "gender").
		Where("uid = ?", member.Uid).
		Update(member)
	return aff == 1, err
}

func (s *Store) SetUserBlockade(uid string, is bool) (bool, error) {
	return s.setUserPermission(uid, "is_blockade", is)
}

func (s *Store) SetUserBanned(uid string, is bool) (bool, error) {
	return s.setUserPermission(uid, "is_message", is)
}

func (s *Store) setUserPermission(uid, colName string, is bool) (bool, error) {
	//用uid找出聊天室中的 member_id
	m, err := s.Find(uid)
	if err != nil {
		return false, err
	}

	log.Debug("models/member.go setUserPermission",
		zap.String("uid", uid),
		zap.String("columnName", colName),
		zap.Bool("bool", is),
		zap.String("Name", m.Name),
		zap.Int64("id", m.Id))

	//TODO: suspection
	/*  // update members table
	aff, err := s.d.Cols(colName).
		Where("id = ?", m.Id).
		Update(&Member{
			IsMessage:  !is,
			IsBlockade: is,
		})
	*/

	aff, err := s.d.Cols(colName).
		Where("uid = ?", uid).
		Update(&Member{
			IsMessage:  is,
			IsBlockade: is,
		})
		//		affected, err := s.d.Exec("UPDATE room_user_permissions SET is_banned=0 , is_blockade=0 WHERE member_id = ?", m.Id)
	log.Debug("db setUserPermission affected row", zap.Int64("affectedRow", aff))

	//解禁
	if err == nil && !is {

		k := map[string]string{
			"is_blockade": "is_blockade",
			"is_message":  "is_banned",
		}
		//s.d.Exec("UPDATE room_user_permissions SET is_banned=0 , is_blockade=0 WHERE member_id = ?", m.Id)
		// update room_user_permission
		aff2, err = s.d.Cols(k[colName]).
			Where("member_id = ?", m.Id).
			Update(&Permission{
				IsBanned:   is,
				IsBlockade: is,
			})
		log.Debug("db setUserPermission aff2", zap.Int64("affectedRow", aff2))
	}

	return aff == 1, err
}

// 找會員
func (s *Store) Find(uid string) (*Member, error) {
	m := new(Member)
	ok, err := s.d.Where("uid = ?", uid).
		Get(m)
	if !ok {
		return nil, sql.ErrNoRows
	}

	return m, err
}

func (s *Store) RoomPermission(id int64, rid int) (Permission, error) {
	p := Permission{}
	_, err := s.d.Where("room_id = ?", rid).Where("member_id = ?", id).Get(&p)
	if err != nil {
		if err == sql.ErrNoRows {
			return Permission{}, nil
		}
		return Permission{}, err
	}

	return p, nil
}

func (s *Store) SetRoomPermission(member Member) error {
	data := &member.Permission

	ok, err := s.d.Where("room_id = ?", member.Permission.RoomId).
		Where("member_id = ?", member.Id).
		Exist(data)
	if err != nil {
		return err
	}

	if ok {
		_, err = s.d.Cols("is_banned", "is_blockade", "is_manage").
			Where("room_id = ?", member.Permission.RoomId).
			Where("member_id = ?", member.Id).
			Update(data)
	} else {
		data.RoomId = int64(member.Permission.RoomId)
		data.MemberId = member.Id
		_, err = s.d.InsertOne(data)
	}

	return err
}

func (s *Store) GetMembers(ids []int64) ([]Member, error) {
	m := make([]Member, 0)
	err := s.d.Table(&Member{}).In("id", ids).Find(&m)
	return m, err
}

func (s *Store) GetMembersByUid(uid []string) ([]Member, error) {
	m := make([]Member, 0)
	err := s.d.In("uid", uid).Find(&m)
	return m, err
}
