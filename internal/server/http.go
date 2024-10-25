package server

import (
	"nps-auth/configs"
	"nps-auth/pkg/cert"
	"nps-auth/pkg/http"
	npsapi "nps-auth/pkg/nps_api"
	"nps-auth/pkg/sql"
	"reflect"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	AuthKey = "ihouqi20220123456"
)

func initHttp() http.ApiService {
	conf := configs.GetConfig()

	router := newRouter()
	http.MustInitApiService(conf.Http.ServerAddr, router)

	server := http.GetAPIService()
	return server
}

func newRouter() *gin.Engine {
	router := http.NewRouter()

	router.GET("/nps/check", http.MakeGinHandlerFunc(handleCheckAuth))
	router.POST("/nps/signature", http.MakeGinHandlerFunc(handleAuth))
	router.DELETE("/nps/delete", http.MakeGinHandlerFunc(handleDelete))

	router.Any("/proxy/:channel/*proxyParts", dynamicReverseProxy())

	return router
}

func handleCheckAuth(c *gin.Context) (*http.Result, error) {
	// bind query
	type Query struct {
		ChannelID string `form:"channelId"`
		NpsId     string `form:"npsId"`
	}
	var query Query

	err := c.ShouldBindQuery(&query)
	if err != nil {
		log.Error().Err(err).Msg("failed to bind query")
		return http.Err("param error"), nil
	}

	type Response struct {
		Able bool `json:"able"`
	}
	var resp Response = Response{
		Able: false,
	}

	// db.find
	var channel sql.Channel
	sql.GetDB().First(&channel, "channel_id = ?", query.ChannelID)

	// check expired
	if channel.ExpiredTime.Before(time.Now()) {
		log.Error().Interface("channel", channel).Msg("channel expired")
		return http.OK("").WithData(resp), nil
	}

	// 查询tunnel信息
	tunnelId, err := strconv.Atoi(channel.NpsTunnelId)
	if err != nil {
		log.Error().Err(err).Interface("channel", channel).Msg("failed to convert tunnel id")
		return http.OK("").WithData(resp), nil
	}

	tunnelInfo, err := npsApi.GetOneTunnel(npsapi.GetOneTunnelReq{
		ID: tunnelId,
	})
	if err != nil {
		log.Error().Err(err).Interface("channel", channel).Msg("failed to get tunnel info")
		return http.OK("").WithData(resp), nil
	}

	if !tunnelInfo.AjaxOne.Data.Status || !tunnelInfo.AjaxOne.Data.Client.Status {
		log.Error().Interface("tunnelInfo", tunnelInfo).Msg("status is false")
		return http.OK("").WithData(resp), nil
	}

	resp.Able = true
	return http.OK("").WithData(resp), nil
}

func handleAuth(c *gin.Context) (*http.Result, error) {
	type HandleAuthReq struct {
		Key         string `json:"key" binding:"required"` //校验key 防止外部调用
		Csr         string `json:"csr" binding:"required"`
		ExpiredTime int    `json:"expiredTime" binding:"required"` //unix时间戳
		Desc        string `json:"desc" binding:"required"`
	}
	var req = HandleAuthReq{}

	if err := c.ShouldBind(&req); err != nil {
		log.Error().Msg("param error")

		return http.Err("param error"), nil
	}

	if req.Key != AuthKey {
		return http.Err("key is illegal"), nil
	}

	csrData, err := cert.ParseCSR([]byte(req.Csr))

	if err != nil || csrData.MachineId == "" {
		return http.Err("csr is illegal"), nil
	}
	// 生成client
	var verifyKey = csrData.MachineId + "-" + strconv.FormatInt(time.Now().Unix(), 10)
	addClientresp, err := npsApi.AddClient(
		npsapi.AddClientReq{
			Remark:          req.Desc,
			VerifyKey:       verifyKey,
			ConfigConnAllow: 0,
			Compress:        1,
			Crypt:           1,
		},
	)
	if err != nil || addClientresp.AjaxResp.Status == 0 {
		log.Error().Err(err).Interface("addClientresp", addClientresp).Msg("gen client error")
		return http.Err("gen client error"), nil
	}

	// 生成tunnel
	resp, err := npsApi.AddTunnel(
		npsapi.AddTunnelReq{
			Type:     "tcp",
			Remark:   req.Desc,
			Target:   configs.GetConfig().Nps.ClientPort,
			ClientID: strconv.Itoa(addClientresp.ID),
		},
	)
	if err != nil || resp.AjaxResp.Status == 0 {
		log.Error().Err(err).Interface("getOneTunneResp", resp).Msg("gen tunnel error")
		return http.Err("gen tunnel error"), nil
	}
	// 查tunnel 信息
	getOneTunneResp, err := npsApi.GetOneTunnel(
		npsapi.GetOneTunnelReq{
			ID: resp.ID,
		},
	)
	if err != nil || getOneTunneResp.AjaxOne.Code == 0 {
		log.Error().Err(err).Interface("getOneTunneResp", getOneTunneResp).Msg("get tunnel error")
		return http.Err("get tunnel error"), nil
	}

	// 写db
	var channelEntity = sql.ChannelEntity{
		// ChannelId:     0, // 让数据库自动生成
		Desc: req.Desc,

		NpsHost:       configs.GetConfig().Nps.BridgeHost,
		NpsClientId:   strconv.Itoa(addClientresp.ID),
		NpsClientKey:  verifyKey,
		NpsTunnelId:   strconv.Itoa(resp.ID),
		NpsTunnelPort: getOneTunneResp.AjaxOne.Data.Port,

		MachineId:   csrData.MachineId,
		ExpiredTime: time.Unix(int64(req.ExpiredTime), 0),
	}
	var channel = sql.Channel{
		//帮我填充
		ChannelEntity: &channelEntity,
	}

	result := sql.GetDB().Create(&channel)
	if result.Error != nil {
		return http.Err("insert db error"), nil
	}

	// 生成cert
	var certData cert.CertData
	ConvertStruct(channelEntity, &certData)

	certByte, err := cert.GenerateCertificate(certData)
	if err != nil {
		log.Error().Err(err).Msg("gen cert error")
		return http.Err("gen cert error"), nil
	}
	var certString = string(certByte)

	log.Info().Str("ret", certString).Msg("gen cert success")

	go handleDeleteByMachineId(csrData.MachineId, channel.ChannelId)

	return http.OK("").WithData(map[string]any{
		"cert": certString,
	}), nil
}

func handleDelete(c *gin.Context) (*http.Result, error) {
	// bind query
	machineId := c.Query("machineId")

	if err := handleDeleteByMachineId(machineId, 0); err != nil {
		log.Error().Err(err).Msg("handleDelete error")
		return http.Err(err.Error()), nil
	} else {
		log.Info().Msg("handleDelete success")
		return http.OK("handleDelete success"), nil
	}

}

/******************************************************************************
*                                Index                                   *
******************************************************************************/

func ConvertStruct(src, dst interface{}) {
	srcVal := reflect.ValueOf(src)
	dstVal := reflect.ValueOf(dst).Elem() // 获取指针指向的值

	for i := 0; i < dstVal.NumField(); i++ {
		dstField := dstVal.Type().Field(i)
		srcField := srcVal.FieldByName(dstField.Name)
		if srcField.IsValid() && srcField.Type() == dstField.Type {
			dstVal.Field(i).Set(srcField)
		}
	}
}

func handleDeleteByMachineId(machineId string, ignoreChannelId int) error {
	var channleList []sql.Channel
	if err := sql.GetDB().Where("machine_id = ?", machineId).Find(&channleList).Error; err != nil {
		log.Error().Err(err).Msg("query channel error")
		return err
	}

	deletedIDs := make([]int, 0, len(channleList))
	for _, channel := range channleList {
		if channel.ChannelId == ignoreChannelId {
			continue
		}

		_, err := npsApi.DeleteClient(npsapi.DeleteClientReq{
			ID: channel.NpsClientId,
		})
		if err != nil {
			log.Error().Err(err).Interface("channel", channel).Msg("delet channel error")
		} else {
			deletedIDs = append(deletedIDs, channel.ChannelId)
		}
	}
	// 从sql 删除
	if err := sql.GetDB().Where("channel_id IN ?", deletedIDs).Delete(&channleList, deletedIDs).Error; err != nil {
		log.Error().Err(err).Msg("delet channel error")
		return err
	}
	log.Info().Interface("deletedIDs", deletedIDs).Msg("delet channel success")

	return nil

}
