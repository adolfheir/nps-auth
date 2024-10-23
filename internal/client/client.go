package client

import (
	logger "nps-auth/pkg/logger"

	"github.com/rs/zerolog"
)

var (
	log zerolog.Logger = logger.GetLogger("client")
	npc *Npc
)

func Init() {
	npc = initNpc()

	server := initHttp()

	server.Run()

}
