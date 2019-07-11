package logic

import (
	"gitlab.com/jetfueltw/cpw/alakazam/client"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/cache"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/conf"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/stream"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"gitlab.com/jetfueltw/cpw/micro/database"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"gitlab.com/jetfueltw/cpw/micro/redis"
	"time"
)

const (
	onlineTick = time.Second * 30
)

// Logic struct
type Logic struct {
	//
	c *conf.Config

	db *models.Store

	cache *cache.Cache

	stream *stream.Stream

	client *client.Client

	// 房間在線人數，key是房間id
	roomCount map[string]int32
}

// New init
func New(c *conf.Config) (l *Logic) {
	l = &Logic{
		c:      c,
		db:     models.NewStore(c.DB),
		cache:  cache.NewRedis(c.Redis),
		stream: stream.NewKafkaPub(c.Kafka),
		client: client.New(c.Api),
	}
	_ = l.loadOnline()
	go l.onlineproc()
	return l
}

func NewAdmin(c1 *database.Conf, c2 *redis.Conf, c3 *conf.Kafka) (*Logic) {
	return &Logic{
		db:     models.NewStore(c1),
		cache:  cache.NewRedis(c2),
		stream: stream.NewKafkaPub(c3),
	}
}

// Close close resources.
func (l *Logic) Close() {
	if err := l.cache.Close(); err != nil {
		log.Errorf("logic close error(%v)", err)
	}
	log.Infof("logic close")
}

func (l *Logic) onlineproc() {
	for {
		time.Sleep(onlineTick)
		if err := l.loadOnline(); err != nil {
			log.Errorf("onlineproc error(%v)", err)
		}
	}
}

// 從redis拿出現在各房間人數
func (l *Logic) loadOnline() (err error) {
	var (
		roomCount = make(map[string]int32)
	)
	var online *cache.Online
	// TODO hostname 先寫死 後續需要註冊中心來sync
	online, err = l.cache.ServerOnline("hostname")
	if err != nil {
		return
	}

	for roomID, count := range online.RoomCount {
		roomCount[roomID] += count
	}
	l.roomCount = roomCount
	return
}
