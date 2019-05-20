package logic

import (
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic/store"
	"os"
	"time"

	log "github.com/golang/glog"
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic/conf"
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic/dao"
)

const (
	onlineTick = time.Second * 30
)

// Logic struct
type Logic struct {
	//
	c *conf.Config

	db *store.Store

	cache *dao.Cache

	stream *dao.Stream

	// 房間在線人數，key是房間id
	roomCount map[string]int32
}

// New init
func New(c *conf.Config) (l *Logic) {
	l = &Logic{
		c:      c,
		db:     store.NewStore(c.DB),
		cache:  dao.NewRedis(c.Redis),
		stream: dao.NewKafkaPub(c.Kafka),
	}
	_ = l.loadOnline()
	go l.onlineproc()
	return l
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
	host, _ := os.Hostname()
	var online *dao.Online
	online, err = l.cache.ServerOnline(host)
	if err != nil {
		return
	}

	for roomID, count := range online.RoomCount {
		roomCount[roomID] += count
	}
	l.roomCount = roomCount
	return
}
