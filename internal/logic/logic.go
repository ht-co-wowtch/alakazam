package logic

import (
	"context"
	"os"
	"time"

	"gitlab.com/jetfueltw/cpw/alakazam/internal/logic/conf"
	"gitlab.com/jetfueltw/cpw/alakazam/internal/logic/dao"
	"gitlab.com/jetfueltw/cpw/alakazam/internal/logic/model"
	log "github.com/golang/glog"
)

const (
	_onlineTick     = time.Second * 10
	_onlineDeadline = time.Minute * 5
)

// Logic struct
type Logic struct {
	//
	c *conf.Config

	// kafka and redis Dao
	dao *dao.Dao

	// 房間在線人數，key是房間id
	roomCount map[string]int32
}

// New init
func New(c *conf.Config) (l *Logic) {
	l = &Logic{
		c:   c,
		dao: dao.New(c),
	}
	_ = l.loadOnline()
	go l.onlineproc()
	return l
}

// Ping ping resources is ok.
func (l *Logic) Ping(c context.Context) (err error) {
	return l.dao.Ping(c)
}

// Close close resources.
func (l *Logic) Close() {
	l.dao.Close()
}

func (l *Logic) onlineproc() {
	for {
		time.Sleep(_onlineTick)
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
	var online *model.Online
	online, err = l.dao.ServerOnline(context.Background(), host)
	if err != nil {
		return
	}
	if time.Since(time.Unix(online.Updated, 0)) > _onlineDeadline {
		_ = l.dao.DelServerOnline(context.Background(), host)
	}
	for roomID, count := range online.RoomCount {
		roomCount[roomID] += count
	}
	l.roomCount = roomCount
	return
}
