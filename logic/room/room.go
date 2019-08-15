package room

import (
	"gitlab.com/jetfueltw/cpw/alakazam/logic/cache"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/member"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
)

type Room struct {
	db *models.Store

	cache *cache.Cache

	member *member.Member
}

func New(db *models.Store, cache *cache.Cache, member *member.Member) *Room {
	return &Room{
		db:     db,
		cache:  cache,
		member: member,
	}
}


