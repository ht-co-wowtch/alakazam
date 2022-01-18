module gitlab.com/ht-co/wowtch/live/alakazam

go 1.13

require (
	github.com/Shopify/sarama v1.23.0
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/eapache/go-resiliency v1.2.0 // indirect
	github.com/gin-contrib/cors v1.3.0
	github.com/gin-gonic/gin v1.4.0
	github.com/go-redis/redis v6.15.2+incompatible
	github.com/go-sql-driver/mysql v1.4.1
	github.com/go-xorm/xorm v0.7.3
	github.com/gogo/protobuf v1.2.1
	github.com/golang/protobuf v1.3.2
	github.com/google/uuid v1.1.1
	github.com/onsi/ginkgo v1.8.0 // indirect
	github.com/onsi/gomega v1.5.0 // indirect
	github.com/prometheus/client_golang v1.1.0
	github.com/rcrowley/go-metrics v0.0.0-20190706150252-9beb055b7962
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.3.0
	github.com/zhenjl/cityhash v0.0.0-20131128155616-cdd6a94144ab
	gitlab.com/ht-co/micro v0.23.4 // indirect
	go.uber.org/zap v1.10.0
	golang.org/x/net v0.0.0-20190613194153-d28f0bde5980
	google.golang.org/grpc v1.21.1
	gopkg.in/go-playground/validator.v8 v8.18.2
	gopkg.in/jcmturner/goidentity.v3 v3.0.0 // indirect
)

replace gitlab.com/ht-co/micro => gitlab.com/ht-co/micro.git v0.23.4
