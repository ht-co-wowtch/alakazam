package run

import (
	"gitlab.com/jetfueltw/cpw/alakazam/admin/conf"
	"gitlab.com/jetfueltw/cpw/alakazam/logic"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/http"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/http/admin"
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
