module gitlab.com/jetfueltw/cpw/alakazam

require (
	github.com/Shopify/sarama v1.23.0
	github.com/alicebob/gopher-json v0.0.0-20180125190556-5a6b3ba71ee6 // indirect
	github.com/alicebob/miniredis v2.5.0+incompatible
	github.com/eapache/go-resiliency v1.2.0 // indirect
	github.com/gin-contrib/cors v1.3.0
	github.com/gin-gonic/gin v1.4.0
	github.com/go-redis/redis v6.15.2+incompatible
	github.com/go-sql-driver/mysql v1.4.1
	github.com/go-xorm/core v0.6.2
	github.com/go-xorm/xorm v0.7.3
	github.com/gogo/protobuf v1.2.1
	github.com/golang/protobuf v1.3.1
	github.com/gomodule/redigo v2.0.0+incompatible // indirect
	github.com/google/uuid v1.1.1
	github.com/mattn/go-sqlite3 v1.10.0
	github.com/onsi/ginkgo v1.8.0 // indirect
	github.com/onsi/gomega v1.5.0 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20190706150252-9beb055b7962 // indirect
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.3.0
	github.com/yuin/gopher-lua v0.0.0-20190514113301-1cd887cd7036 // indirect
	github.com/zhenjl/cityhash v0.0.0-20131128155616-cdd6a94144ab
	gitlab.com/jetfueltw/cpw/micro v0.14.1
	go.uber.org/zap v1.10.0
	golang.org/x/net v0.0.0-20190522155817-f3200d17e092
	google.golang.org/grpc v1.21.1
	gopkg.in/go-playground/validator.v8 v8.18.2
	gopkg.in/testfixtures.v2 v2.5.3
)

replace gitlab.com/jetfueltw/cpw/micro => gitlab.com/jetfueltw/cpw/micro.git v0.14.1
