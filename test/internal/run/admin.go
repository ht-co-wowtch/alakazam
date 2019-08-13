package run

import (
	"gitlab.com/jetfueltw/cpw/alakazam/admin/conf"
	admin "gitlab.com/jetfueltw/cpw/alakazam/admin/api"
	"gitlab.com/jetfueltw/cpw/alakazam/logic"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/http"
)

func RunAdmin(path string) func() {
	if err := conf.Read(path + "/admin.yml"); err != nil {
		panic(err)
	}

	srv := logic.NewAdmin(conf.Conf.DB, conf.Conf.Redis, conf.Conf.Kafka)
	httpAdminSrv := http.New(conf.Conf.HTTPServer, admin.New(srv))

	return func() {
		srv.Close()
		httpAdminSrv.Close()
	}
}
