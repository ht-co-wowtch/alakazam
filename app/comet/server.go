package comet

import (
	"context"
	"math/rand"
	"time"

	"github.com/zhenjl/cityhash"
	"gitlab.com/jetfueltw/cpw/alakazam/app/comet/conf"
	"gitlab.com/jetfueltw/cpw/alakazam/app/logic/pb"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"gitlab.com/jetfueltw/cpw/micro/grpc"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
)

const (
	version = "v1.83.1"
	// 通知logic Refresh client連線狀態最小心跳時間
	minServerHeartbeat = time.Minute * 10

	// 通知logic Refresh client連線狀態最大心跳時間
	maxServerHeartbeat = time.Minute * 20
)

func newLogicClient(c *grpc.Conf) pb.LogicClient {
	conn, err := grpc.NewClient(c)
	if err != nil {
		panic(err)
	}
	return pb.NewLogicClient(conn)
}

// comet server
type Server struct {
	c *conf.Config

	// 控管Reader and Writer Buffer 與Timer的Pool
	round *Round

	// 管理buckets，各紀錄部分的Channel與Room
	buckets []*Bucket

	// buckets總數
	bucketIdx uint32

	// 此comet服務名稱，在分佈式下可能會有多組comet server
	// 用name來區別各個comet server讓job可以準確推送訊息到某user所在的comet server
	name string

	// Logic service grpc client
	logic pb.LogicClient

	// 房間總人數(非即時)
	online map[int32]int32
}

// new Server
func NewServer(c *conf.Config) *Server {
	s := &Server{
		c:     c,
		round: NewRound(c),
		logic: newLogicClient(c.Logic),
	}

	// 初始化Bucket
	s.buckets = make([]*Bucket, c.Bucket.Size)
	s.bucketIdx = uint32(c.Bucket.Size)
	for i := 0; i < c.Bucket.Size; i++ {
		s.buckets[i] = NewBucket(c.Bucket)
	}

	// TODO hostname 先寫死 後續需要註冊中心來sync
	// 坑: 底下的 hostname字串會被用於 room/room.go - Online method中
	s.name = "hostname"

	go s.KickClosedRoomUserPeriod(models.NewStore(c.DB))

	// 統計線上各房間人數
	go s.onlineproc()
	return s
}

func (s *Server) KickClosedRoomUserPeriod(store *models.Store) {
	log.Info("Buckets", zap.Any("Buckets", s.buckets))
	var (
		closedRoomIds []int
		err           error
		counter       int
	)
	for {
		time.Sleep(time.Second * 10)
		log.Info("KickClosedRoomUserPeriod", zap.Int("", counter), zap.String("versioni", version))
		closedRoomIds, err = store.GetClosedRoomIds()
		if err != nil {
			log.Error("[server.go]KickClosedRoomUserPeriod", zap.Error(err))
			return
		}
		if len(closedRoomIds) > 0 {
			log.Info("關閉的聊天室房間(RoomIds)", zap.Ints("roomIds", closedRoomIds))
		}
		//Caution: No Lock here
		for _, roomId := range closedRoomIds {
			for _, bkt := range s.buckets {
				bkt.DelClosedRoom(roomId)
			}
		}
		counter++
	}
}

// 所有buckets
func (s *Server) Buckets() []*Bucket {
	return s.buckets
}

// 根據user key 採用CityHash32算法除於bucket總數的出來的餘數，來取出某個bucket
// 用意在同時間針對不同房間做推播時可以避免使用到同一把鎖，降低鎖的競爭
// 所以可以調高bucket來增加併發量，但同時會多佔用內存
func (s *Server) Bucket(subKey string) *Bucket {
	return s.buckets[(cityhash.CityHash32([]byte(subKey), uint32(len(subKey))) % s.bucketIdx)]
}

// 通知logic Refresh client連線狀態的時間(心跳包的週期)
// 這邊使用隨機產生時間是為了不讓用戶都在同一時間做心跳，以達到分散尖峰
func (s *Server) RandServerHearbeat() time.Duration {
	return (minServerHeartbeat + time.Duration(rand.Int63n(int64(maxServerHeartbeat-minServerHeartbeat))))
}

func (s *Server) Close() (err error) {
	return
}

// 統計各房間人數並發給logic service做更新
// 因為logic有提供http獲得某房間人數
func (s *Server) onlineproc() {
	for {
		var err error
		roomCount := make(map[int32]int32)

		// 因為房間會分散在不同的bucket所以需要統計
		for _, bucket := range s.buckets {
			for roomID, count := range bucket.RoomsCount() {
				roomCount[roomID] += count
			}
		}

		if s.online, err = s.RenewOnline(context.Background(), s.name, roomCount); err != nil {
			time.Sleep(time.Second)
			continue
		}
		for _, bucket := range s.buckets {
			bucket.UpRoomsCount(s.online)
		}

		// 每30秒統計一次發給logic
		time.Sleep(time.Second * 30)
	}
}
