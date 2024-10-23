package client

import (
	"nps-auth/configs"
	"nps-auth/pkg/cert"
	"nps-auth/pkg/http"
	"nps-auth/pkg/processManager"

	"github.com/gin-gonic/gin"
)

func initHttp() http.ApiService {

	conf := configs.GetConfig()

	router := newRouter()
	http.MustInitApiService(conf.Http.ClientAddr, router)

	server := http.GetAPIService()
	return server
}

func newRouter() *gin.Engine {
	router := gin.Default()

	router.GET("/npc/check", http.MakeGinHandlerFunc(handleCheck))
	router.GET("/npc/csr", http.MakeGinHandlerFunc(hanldeGetCsr))
	router.POST("/npc/auth", http.MakeGinHandlerFunc(handlePostCert))
	router.POST("/npc/start", http.MakeGinHandlerFunc(handleStart))
	router.POST("/npc/stop", http.MakeGinHandlerFunc(handleStop))

	return router
}

func handleCheck(ctx *gin.Context) (*http.Result, error) {

	var status processManager.ProcessState

	if npc.process != nil {
		status = npc.process.GetStatus()
	} else {
		status = processManager.Stopped
	}

	// todo: 通过接口尝试连接代理地址, 判断是否运行中

	var data = map[string]any{
		"cert":      npc.cert,
		"certData":  npc.certData,
		"npcStatus": status,
	}

	return http.OK("").WithData(data), nil
}

func hanldeGetCsr(ctx *gin.Context) (*http.Result, error) {
	certByte, err := cert.GenerateCSR()

	if err != nil {
		log.Error().Err(err).Msg("generate csr err")
		return nil, &http.HTTPError{Code: 500, Err: err}
	}

	csr := string(certByte)
	var data = map[string]any{
		"csr": csr,
	}

	return http.OK("").WithData(data), nil
}

func handlePostCert(c *gin.Context) (*http.Result, error) {

	var req struct {
		Cert string `json:"cert"`
	}
	if err := c.ShouldBind(&req); err != nil {
		log.Error().Msg("param error")
		return http.Err("param error"), nil
	}
	// 保存 cert 到本地文件
	err := SaveCertToFile(req.Cert)
	if err != nil {
		log.Error().Err(err).Msg("save cert err")
		return nil, &http.HTTPError{Code: 500, Err: err}
	}

	// 重新加载 cert
	err = npc.LoadCert(req.Cert)
	if err != nil {
		log.Error().Err(err).Msg("load cert err")
		return nil, &http.HTTPError{Code: 500, Err: err}
	}

	return http.OK(""), nil
}

func handleStart(ctx *gin.Context) (*http.Result, error) {
	err := npc.StartNpc()

	if err != nil {
		log.Error().Err(err).Msg("start npc err")
		return nil, &http.HTTPError{Code: 500, Err: err}
	}

	return http.OK("start success"), nil
}

func handleStop(ctx *gin.Context) (*http.Result, error) {
	err := npc.StopNpc()

	if err != nil {
		log.Error().Err(err).Msg("stop npc err")
		return nil, &http.HTTPError{Code: 500, Err: err}
	}

	return http.OK("stop success"), nil
}
