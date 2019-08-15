package logic

import (
	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
)

// 登入到聊天室
func (l *Logic) login(token, roomId, server string) (*models.Member, string, error) {
	user, err := l.client.Auth(token)
	if err != nil {
		return nil, "", err
	}

	u, ok, err := l.db.Find(user.Uid)
	if err != nil {
		return nil, "", err
	}
	if !ok {
		u = &models.Member{
			Uid:    user.Uid,
			Name:   user.Name,
			Avatar: user.Avatar,
			Type:   user.Type,
		}
		if aff, err := l.db.CreateUser(u); err != nil || aff <= 0 {
			return nil, "", err
		}
	} else if u.IsBlockade {
		return u, "", nil
	}

	if u.Name != user.Name || u.Avatar != user.Avatar {
		u.Name = user.Name
		u.Avatar = user.Avatar
		if aff, err := l.db.UpdateUser(u); err != nil || aff <= 0 {
			log.Error("UpdateUser", zap.String("uid", user.Uid), zap.Int64("affected", aff), zap.Error(err))
		}
	}

	key := uuid.New().String()

	// 儲存user資料至redis
	if err := l.cache.SetUser(u, key, roomId, server); err != nil {
		return nil, "", err
	} else {
		log.Info(
			"conn connected",
			zap.String("key", key),
			zap.String("uid", u.Uid),
			zap.String("room_id", roomId),
			zap.String("server", server),
		)
	}
	return u, key, nil
}

func (l *Logic) GetUserName(uid []string) ([]string, error) {
	name, err := l.cache.GetUserName(uid)
	if err == redis.Nil {
		return nil, errors.ErrNoRows
	}
	return name, nil
}
