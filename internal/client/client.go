package client

import (
	"nps-auth/pkg/cert"
	logger "nps-auth/pkg/logger"

	"github.com/rs/zerolog"
)

var (
	log zerolog.Logger = logger.GetLogger("client")
	npc *Npc
)

func Init() {
	// 获取机器码比较慢 这边做一次预加载
	go cert.GetMachineID()

	npc = initNpc()

	server := initHttp()
	server.Run()

}
