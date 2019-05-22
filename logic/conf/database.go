package conf

import (
	"github.com/spf13/viper"
	"os"
	"time"
)

// Database 相關設定
type Database struct {
	Driver    string
	Host      string
	Port      string
	Database  string
	User      string
	Password  string
	Charset   string
	Collation string

	// 最大連線總數
	MaxOpenConn int

	// 最大保留的閒置連線數
	MaxIdleConn int

	// 空閒連線多久(秒)沒做事就close
	ConnMaxLifetime time.Duration
}

func newDatabase() *Database {
	db := new(Database)
	db.Driver = "mysql"
	db.Host = viper.GetString("db.host")
	db.Port = viper.GetString("db.port")
	db.Database = viper.GetString("db.database")
	db.Charset = viper.GetString("db.charset")
	db.Collation = viper.GetString("db.collation")
	db.MaxOpenConn = viper.GetInt("db.active")
	db.MaxIdleConn = viper.GetInt("db.idle")
	db.MaxIdleConn = viper.GetInt("db.idleTimeout")

	if u := viper.GetString("db.user"); u != "" {
		db.User = u
	} else {
		db.User = os.Getenv("DB_USER")
	}

	if p := viper.GetString("db.password"); p != "" {
		db.Password = p
	} else {
		db.Password = os.Getenv("DB_PASSWORD")
	}
	return db
}
