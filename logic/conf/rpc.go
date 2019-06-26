package conf

import (
	"github.com/spf13/viper"
	"gitlab.com/jetfueltw/cpw/micro/grpc"
)

func newRpc() *grpc.Conf {
	// TODO 處理error
	g, _ := grpc.ReadViper(viper.Sub("grpcServer"))
	return g
}
