package room

import (
	"github.com/go-redis/redis"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/member"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
)

func (r *Room) Auth(u *member.User) error {
	var err error
	u.RoomStatus, err = r.c.get(u.H.Room)
	if err != nil && err != redis.Nil {
		return err
	}
	if u.RoomStatus == 0 {
		room, _, err := r.db.GetRoom(u.H.Room)
		if err != nil {
			return err
		}
		u.RoomStatus = room.Permission()
		if err := r.c.set(room); err != nil {
			return err
		}
	}
	if models.IsBanned(u.RoomStatus) {
		return errors.ErrRoomBanned
	}
	if !u.H.IsMessage {
		return errors.ErrBanned
	}
	is, err := r.member.IsBanned(u.Uid)
	if err != nil {
		return err
	}
	if is {
		return errors.ErrBanned
	}
	return nil
}
