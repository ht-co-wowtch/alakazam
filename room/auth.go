package room

import (
	"github.com/go-redis/redis"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"strconv"
)

func (r *Room) GetUserRoom(uid, rId string) (*models.Room, error) {
	room, err := r.c.get(rId)
	if room == nil {
		if err != redis.Nil {
			return nil, err
		}

		id, err := strconv.Atoi(rId)
		if err != nil {
			return nil, err
		}

		room, ok, err := r.db.GetRoom(id)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, errors.ErrNoRows
		}
		if err := r.c.set(room); err != nil {
			log.Error("set room cache", zap.Error(err))
		}
	}
	return room, nil
}
