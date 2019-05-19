package logic

import (
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

	// kafka and redis Dao
	dao *dao.Dao

	db *dao.Store

	cache *dao.Cache

	// 房間在線人數，key是房間id
	roomCount map[string]int32
}

// New init
func New(c *conf.Config) (l *Logic) {
	l = &Logic{
		c:     c,
		dao:   dao.New(c),
		db:    dao.NewStore(c.DB),
		cache: dao.NewRedis(c.Redis),
	}
	_ = l.loadOnline()
	go l.onlineproc()
	return l
}

// Close close resources.
func (l *Logic) Close() {
	l.cache.Close()
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
