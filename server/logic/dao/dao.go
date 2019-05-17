package dao

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gomodule/redigo/redis"
	_ "github.com/mattn/go-sqlite3"
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic/conf"
	kafka "gopkg.in/Shopify/sarama.v1"
	"time"
)

type Dao struct {
	c        *conf.Config
	kafkaPub kafka.SyncProducer
	redis    *redis.Pool
	db       *sql.DB

	// redis 過期時間
	redisExpire int32
}

func New(c *conf.Config) *Dao {
	d := &Dao{
		c:           c,
		db:          newDB(c.DB),
		kafkaPub:    newKafkaPub(c.Kafka),
		redis:       newRedis(c.Redis),
		redisExpire: int32(c.Redis.Expire / time.Second),
	}
	return d
}

func newKafkaPub(c *conf.Kafka) kafka.SyncProducer {
	kc := kafka.NewConfig()
	kc.Producer.RequiredAcks = kafka.WaitForAll
	kc.Producer.Retry.Max = 10
	kc.Producer.Return.Successes = true
	pub, err := kafka.NewSyncProducer(c.Brokers, kc)
	if err != nil {
		panic(err)
	}
	return pub
}

func newRedis(c *conf.Redis) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     c.Idle,
		MaxActive:   c.Active,
		IdleTimeout: c.IdleTimeout,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial(c.Network, c.Addr,
				redis.DialConnectTimeout(c.DialTimeout),
				redis.DialReadTimeout(c.ReadTimeout),
				redis.DialWriteTimeout(c.WriteTimeout),
			)
			if err != nil {
				return nil, err
			}
			return conn, nil
		},
	}
}

func newDB(c *conf.Database) *sql.DB {
	db, err := sql.Open(c.Driver, DatabaseDns(c))
	if err != nil {
		panic(err)
	}

	db.SetMaxOpenConns(c.MaxOpenConn)
	db.SetMaxIdleConns(c.MaxIdleConn)
	db.SetConnMaxLifetime(time.Duration(c.ConnMaxLifetime) * time.Second)

	if err := db.Ping(); err != nil {
		panic(err)
	}
	return db
}

func DatabaseDns(c *conf.Database) string {
	return fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=%v&collation=%v&parseTime=true&timeout=2s&loc=Local", c.User, c.Password, c.Host, c.Port, c.Database, c.Charset, c.Collation)
}

// Close close the resource.
func (d *Dao) Close() error {
	return d.redis.Close()
}

// ping redis是否活著
func (d *Dao) Ping() error {
	return d.pingRedis()
}
