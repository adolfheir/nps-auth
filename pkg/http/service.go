package http

import (
	"nps-auth/configs"
	"nps-auth/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type ApiService interface {
	Run()
	Stop()
}

type server struct {
	srv  *gin.Engine
	addr string
	log  zerolog.Logger // 添加 logger 字段
}

var (
	_          ApiService = (*server)(nil)
	apiService *server
	ginLog     zerolog.Logger = logger.GetLogger("gin")
)

func NewRouter() *gin.Engine {
	conf := configs.GetConfig()

	r := gin.New()

	if conf.Logger.Output == "file" {
		gin.DisableConsoleColor()
		writer := logger.GetFileWriter()
		gin.DefaultWriter = writer
	}

	return r
}

func NewApiService(listenAddr string, router *gin.Engine) *server {
	return &server{
		srv:  router,
		addr: listenAddr,
		log:  ginLog,
	}
}

func MustInitApiService(listenAddr string, router *gin.Engine) {
	apiService = NewApiService(listenAddr, router)
}

func GetAPIService() ApiService {
	return apiService
}

func (s *server) Run() {
	s.log.Info().Msgf("Starting server on %s", s.addr) // 记录启动日志
	if err := s.srv.Run(s.addr); err != nil {
		s.log.Error().Err(err).Msg("Server failed to start") // 记录启动失败日志
	}
}

func (s *server) Stop() {
	s.log.Info().Msg("Stopping server") // 记录停止日志
	// 这里可以添加停止服务器的逻辑
}
