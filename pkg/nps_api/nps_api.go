package npsapi

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"

	"github.com/go-resty/resty/v2"
)

// API 是 API 调用的结构体
type API struct {
	baseURL string
	authKey string
	client  *resty.Client
}

// NewAPI 创建一个新的 API 实例
func NewAPI(baseURL, authKey string) *API {
	return &API{
		baseURL: baseURL,
		authKey: authKey,
		client:  resty.New(),
	}
}

// generateAuthKey 生成 auth_key
func (a *API) generateAuthKey(timestamp int64) string {
	h := md5.New()
	h.Write([]byte(fmt.Sprintf("%s%d", a.authKey, timestamp)))
	return hex.EncodeToString(h.Sum(nil))
}

// getCurrentTimestamp 获取当前时间戳
func (a *API) getCurrentTimestamp() (int64, error) {

	type GetCurrentTimestampResp struct {
		Time int64 `json:"time"`
	}

	data := GetCurrentTimestampResp{}

	_, err := a.client.R().
		SetResult(&data).
		Post(a.baseURL + "/auth/gettime")

	if err != nil {
		return 0, err
	}

	time := data.Time

	return time, nil
}

// postRequest 发送 POST 请求
func (a *API) postRequest(endpoint string, data interface{}) (*resty.Response, error) {

	timestamp, err := a.getCurrentTimestamp()
	if err != nil {
		return nil, err
	}

	authKey := a.generateAuthKey(timestamp)

	formData := url.Values{}
	formData.Set("auth_key", authKey)
	formData.Set("timestamp", fmt.Sprintf("%d", timestamp))

	// 使用反射填充 formData
	// 假设所有的 data 结构体字段都可导出，并为字符串或数字
	for key, value := range structToMap(data) {
		var value = fmt.Sprintf("%v", value)
		if value != "" {
			formData.Set(key, fmt.Sprintf("%v", value))
		}
	}

	resp, err := a.client.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetBody(formData.Encode()).
		Post(a.baseURL + endpoint)

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// structToMap 辅助函数：将结构体转换为 map
func structToMap(v interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	val := reflect.ValueOf(v)

	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" {
			result[jsonTag] = val.Field(i).Interface()
		}
	}
	return result
}

/******************************************************************************
*                                api                                   *
******************************************************************************/

// GetClientListReq 获取客户端列表的参数结构体
type GetClientListReq struct {
	Search string `json:"search"`
	Order  string `json:"order"`
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
}

type GetClientListResp struct {
	BridgePort int    `json:"bridgePort"`
	BridgeType string `json:"bridgeType"`
	IP         string `json:"ip"`
	AjaxTable[Client]
}

// GetClientList 获取客户端列表
func (a *API) GetClientList(params GetClientListReq) (*GetClientListResp, error) {
	resp, err := a.postRequest("/client/list/", params)
	if err != nil {
		return nil, err
	}

	var data GetClientListResp
	marshalErr := json.Unmarshal(resp.Body(), &data)

	if marshalErr != nil {
		return nil, marshalErr
	}

	return &data, nil
}

// GetClientByIDReq 根据 ID 获取客户端的参数结构体
type GetClientByIDReq struct {
	ID string `json:"id"`
}

type GetClientByIDResp struct {
	AjaxOne[Client]
}

// GetClientByID 获取单个客户端
func (a *API) GetClientByID(params GetClientByIDReq) (*GetClientByIDResp, error) {
	resp, err := a.postRequest("/client/list/", params)
	if err != nil {
		return nil, err
	}
	data := resp.Result().(GetClientByIDResp)

	return &data, nil
}

// AddClientReq 添加客户端的参数结构体
type AddClientReq struct {
	Remark          string `json:"remark"`
	VerifyKey       string `json:"vkey"`
	ConfigConnAllow int    `json:"config_conn_allow"` // 是否允许客户端以配置文件模式连接  1允许 0不允许
	Compress        int    `json:"compress"`
	Crypt           int    `json:"crypt"`
}

type AddClientResp struct {
	AjaxResp
	ID int `json:"id"`
}

// AddClient 添加客户端
func (a *API) AddClient(params AddClientReq) (*AddClientResp, error) {

	resp, err := a.postRequest("/client/add/", params)
	if err != nil {
		return nil, err
	}

	var data AddClientResp
	marshalErr := json.Unmarshal(resp.Body(), &data)

	if marshalErr != nil {
		fmt.Printf("marshalErr resp string: %v\n", resp.String())
		return nil, marshalErr
	}

	return &data, nil
}

// EditClientReq 修改客户端的参数结构体
type EditClientReq struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	// 其他需要修改的参数
}

// EditClient 修改客户端
func (a *API) EditClient(params EditClientReq) (interface{}, error) {
	return a.postRequest("/client/edit/", params)
}

// DeleteClientReq 删除客户端的参数结构体
type DeleteClientReq struct {
	ID string `json:"id"`
}

// DeleteClient 删除客户端
func (a *API) DeleteClient(params DeleteClientReq) (interface{}, error) {
	return a.postRequest("/client/del/", params)
}

type GetOneTunnelReq struct {
	ID int `json:"id"`
}
type GetOneTunnelResp struct {
	AjaxOne[Tunnel]
}

// GetOneTunnel 获取单条隧道信息
func (a *API) GetOneTunnel(params GetOneTunnelReq) (*GetOneTunnelResp, error) {

	resp, err := a.postRequest("/index/getonetunnel/", params)
	if err != nil {
		return nil, err
	}

	var data GetOneTunnelResp
	marshalErr := json.Unmarshal(resp.Body(), &data)

	if marshalErr != nil {
		fmt.Printf("marshalErr resp string: %v\n", resp.String())
		return nil, marshalErr
	}

	return &data, nil
}

// GetTunnelListReq 隧道相关的参数结构体
type GetTunnelListReq struct {
	ClientID string `json:"client_id"`
	Type     string `json:"type"`
	Search   string `json:"search"`
	Offset   int    `json:"offset"`
	Limit    int    `json:"limit"`
}

// GetTunnelListResp 隧道列表的返回结构体
type GetTunnelListResp struct {
	AjaxTable[Tunnel]
}

// GetTunnelList 获取隧道列表
func (a *API) GetTunnelList(params GetTunnelListReq) (*GetTunnelListResp, error) {

	resp, err := a.postRequest("/index/gettunnel/", params)
	if err != nil {
		return nil, err
	}

	var data GetTunnelListResp
	marshalErr := json.Unmarshal(resp.Body(), &data)

	if marshalErr != nil {
		return nil, marshalErr
	}

	return &data, nil
}

// AddTunnelReq 添加隧道的参数结构体
type AddTunnelReq struct {
	Type     string `json:"type"`      // 类型: tcp, udp, httpProx, socks5, secret, p2p
	Remark   string `json:"remark"`    // 备注信息
	Port     int    `json:"port"`      // 服务端端口
	Target   string `json:"target"`    // 目标地址 (格式: ip:端口)
	ClientID string `json:"client_id"` // 客户端ID
}
type AddTunnelResp struct {
	AjaxResp
	ID int `json:"id"`
}

// AddTunnel 添加隧道
func (a *API) AddTunnel(params AddTunnelReq) (*AddTunnelResp, error) {
	resp, err := a.postRequest("/index/add/", params)
	if err != nil {
		return nil, err
	}

	var data AddTunnelResp
	marshalErr := json.Unmarshal(resp.Body(), &data)

	if marshalErr != nil {
		return nil, marshalErr
	}

	return &data, nil
}

// EditTunnelReq 修改隧道的参数结构体
type EditTunnelReq struct {
	ID string `json:"id"`
	// 其他需要修改的参数
}

// EditTunnel 修改隧道
func (a *API) EditTunnel(params EditTunnelReq) (interface{}, error) {
	return a.postRequest("/index/edit/", params)
}

// DeleteTunnelReq 删除隧道的参数结构体
type DeleteTunnelReq struct {
	ID string `json:"id"`
}

// DeleteTunnel 删除隧道
func (a *API) DeleteTunnel(params DeleteTunnelReq) (interface{}, error) {
	return a.postRequest("/index/del/", params)
}

// StopTunnelReq 停止隧道的参数结构体
type StopTunnelReq struct {
	ID string `json:"id"`
}

// StopTunnel 停止隧道
func (a *API) StopTunnel(params StopTunnelReq) (*AjaxResp, error) {
	resp, err := a.postRequest("/client/list/", params)
	if err != nil {
		return nil, err
	}

	var data AjaxResp
	marshalErr := json.Unmarshal(resp.Body(), &data)

	if marshalErr != nil {
		return nil, marshalErr
	}

	return &data, nil

}

// StartTunnelReq 启动隧道的参数结构体
type StartTunnelReq struct {
	ID string `json:"id"`
}

// StartTunnel 启动隧道
func (a *API) StartTunnel(params StartTunnelReq) (interface{}, error) {
	return a.postRequest("/index/start/", params)
}
