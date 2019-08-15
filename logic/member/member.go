package member

import (
	"gitlab.com/jetfueltw/cpw/alakazam/logic/cache"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
)

type Member struct {
	db *models.Store

	cache *cache.Cache
}

func New(db *models.Store, cache *cache.Cache) *Member {
	return &Member{
		db:    db,
		cache: cache,
	}
}
