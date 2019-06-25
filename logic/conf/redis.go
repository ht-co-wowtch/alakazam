package conf

import (
	"github.com/spf13/viper"
	"gitlab.com/jetfueltw/cpw/micro/redis"
	"time"
)

// Cache
type Redis struct {
	// host
	Network string

	// port
	Addr string

	// pool內最大連線總數
	Active int

	// 最大保留的閒置連線數
	Idle int

	// 建立連線超時多久後放棄
	DialTimeout time.Duration

	// read多久沒回覆則放棄
	ReadTimeout time.Duration

	// write多久沒回覆則放棄
	WriteTimeout time.Duration

	// 空閒連線多久沒做事就close
	IdleTimeout time.Duration

	// redis過期時間
	Expire time.Duration
}

func newRedis() *redis.Conf {
	v := viper.Sub("redis")

	// TODO error 處理
	c, _ := redis.ReadViper(v)
	return c
}
