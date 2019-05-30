package logic

import (
	"database/sql"
	log "github.com/golang/glog"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/permission"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/store"
)

func (l *Logic) auth(token string) (*store.User, error) {
	user, err := l.client.Auth(token)
	if err != nil {
		log.Errorf("Logic client GetUser token:%s error(%v)", token, err)
		return nil, errors.UserError
	}

	u, err := l.db.Find(user.Uid)

	if err != nil {
		if err != sql.ErrNoRows {
			log.Errorf("FindUserPermission(uid:%s) error(%v)", user.Uid, err)
			return nil, errors.ConnectError
		}

		u = &store.User{
			Uid:        user.Uid,
			Name:       user.Nickname,
			Avatar:     user.Avatar,
			Permission: permission.PlayDefaultPermission,
		}

		if aff, err := l.db.CreateUser(u); err != nil || aff <= 0 {
			log.Errorf("CreateUser(uid:%s) affected %d error(%v)", user.Uid, aff, err)
			return nil, errors.ConnectError
		}
	} else if u.IsBlockade {
		return u, nil
	}

	if u.Name != user.Nickname || u.Avatar != user.Avatar {
		u.Name = user.Nickname
		u.Avatar = user.Avatar
		if aff, err := l.db.UpdateUser(u); err != nil || aff <= 0 {
			log.Errorf("UpdateUser(uid:%s) affected %d error(%v)", user.Uid, aff, err)
		}
	}

	return u, nil
}
