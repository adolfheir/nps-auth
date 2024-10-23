package server

import (
	"nps-auth/pkg/logger"
	npsapi "nps-auth/pkg/nps_api"
	"nps-auth/pkg/sql"

	"github.com/rs/zerolog"
)

var (
	log    zerolog.Logger = logger.GetLogger("server")
	npsApi *npsapi.API
)

func Init() {
	initLru()

	// 初始化npsApi
	npsApi = npsapi.NewAPI("http://175.27.193.51:20100/", "ihouqi2022")

	// 初始化数据库
	sql.GetDB()

	// 初始化http服务
	server := initHttp()
	server.Run()

}
