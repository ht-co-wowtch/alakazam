module gitlab.com/jetfueltw/cpw/alakazam

require (
	github.com/DATA-DOG/go-sqlmock v1.3.3
	github.com/Shopify/sarama v1.22.1 // indirect
	github.com/alicebob/gopher-json v0.0.0-20180125190556-5a6b3ba71ee6 // indirect
	github.com/alicebob/miniredis v2.5.0+incompatible
	github.com/bsm/sarama-cluster v2.1.15+incompatible
	github.com/gin-contrib/cors v1.3.0
	github.com/gin-gonic/gin v1.4.0
	github.com/go-redis/redis v6.15.2+incompatible
	github.com/go-sql-driver/mysql v1.4.1
	github.com/gogo/protobuf v1.2.1
	github.com/golang-migrate/migrate/v4 v4.4.0
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/protobuf v1.3.1
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/google/uuid v1.1.1
	github.com/mattn/go-sqlite3 v1.10.0
	github.com/onsi/ginkgo v1.8.0 // indirect
	github.com/onsi/gomega v1.5.0 // indirect
	github.com/smartystreets/goconvey v0.0.0-20190330032615-68dc04aab96a
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.3.0
	github.com/yuin/gopher-lua v0.0.0-20190514113301-1cd887cd7036 // indirect
	github.com/zhenjl/cityhash v0.0.0-20131128155616-cdd6a94144ab
	gitlab.com/jetfueltw/cpw/micro v0.2.0
	golang.org/x/net v0.0.0-20190522155817-f3200d17e092
	google.golang.org/grpc v1.21.1
	gopkg.in/Shopify/sarama.v1 v1.19.0
	gopkg.in/go-playground/validator.v8 v8.18.2
)

replace gitlab.com/jetfueltw/cpw/micro => gitlab.com/jetfueltw/cpw/micro.git v0.2.0
