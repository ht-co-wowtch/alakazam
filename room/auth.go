package room

import (
	"github.com/go-redis/redis"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/member"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
)

func (r *Room) Auth(u *member.User) error {
	var err error
	u.RoomStatus, err = r.c.get(u.RoomId)
	if err != nil && err != redis.Nil {
		return err
	}
	if u.RoomStatus == 0 {
		room, _, err := r.db.GetRoom(u.RoomId)
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
	is, err := r.member.IsUserBanned(u.Uid, u.Status)
	if err != nil {
		return err
	}
	if is {
		return errors.ErrBanned
	}
	return nil
}
