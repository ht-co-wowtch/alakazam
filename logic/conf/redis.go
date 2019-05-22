package conf

import (
	"github.com/spf13/viper"
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

func newRedis() *Redis {
	return &Redis{
		Network:      "tcp",
		Addr:         viper.GetString("redis.host"),
		Active:       viper.GetInt("redis.active"),
		Idle:         viper.GetInt("redis.idle"),
		DialTimeout:  time.Duration(viper.GetInt("redis.dialTimeout")) * time.Second,
		ReadTimeout:  time.Duration(viper.GetInt("redis.readTimeout")) * time.Second,
		WriteTimeout: time.Duration(viper.GetInt("redis.writeTimeout")) * time.Second,
		IdleTimeout:  time.Duration(viper.GetInt("redis.idleTimeout")) * time.Second,
		Expire:       time.Duration(viper.GetInt("redis.expire")) * time.Second,
	}
}
