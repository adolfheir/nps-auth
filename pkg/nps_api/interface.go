package npsapi

import (
	"time"
)

/******************************************************************************
*                                struct                                   *
******************************************************************************/

type Flow struct {
	ExportFlow int64 `json:"ExportFlow"`
	InletFlow  int64 `json:"InletFlow"`
	FlowLimit  int   `json:"FlowLimit"`
}

type Rate struct {
	NowRate int `json:"NowRate"`
}

// 定义 ClientConfig 结构体
type ClientConfig struct {
	U        string `json:"U"`
	P        string `json:"P"`
	Compress bool   `json:"Compress"`
	Crypt    bool   `json:"Crypt"`
}

// 定义 Client 结构体
type Client struct {
	Cnf             ClientConfig `json:"Cnf"`
	Id              int          `json:"Id"`
	VerifyKey       string       `json:"VerifyKey"`
	Addr            string       `json:"Addr"`
	Remark          string       `json:"Remark"`
	Status          bool         `json:"Status"`
	IsConnect       bool         `json:"IsConnect"`
	RateLimit       int          `json:"RateLimit"`
	Flow            Flow         `json:"Flow"`
	Rate            Rate         `json:"Rate"`
	NoStore         bool         `json:"NoStore"`
	NoDisplay       bool         `json:"NoDisplay"`
	MaxConn         int          `json:"MaxConn"`
	NowConn         int          `json:"NowConn"`
	WebUserName     string       `json:"WebUserName"`
	WebPassword     string       `json:"WebPassword"`
	ConfigConnAllow bool         `json:"ConfigConnAllow"`
	MaxTunnelNum    int          `json:"MaxTunnelNum"`
	Version         string       `json:"Version"`
	BlackIpList     []string     `json:"BlackIpList"`
	LastOnlineTime  string       `json:"LastOnlineTime"`
}

// 定义 Target 结构体
type Target struct {
	TargetStr  string   `json:"TargetStr"`
	TargetArr  []string `json:"TargetArr"`
	LocalProxy bool     `json:"LocalProxy"`
}

type MultiAccount struct {
	AccountMap map[string]string // multi account and pwd
}

type Health struct {
	HealthCheckTimeout  int
	HealthMaxFail       int
	HealthCheckInterval int
	HealthNextTime      time.Time
	HealthMap           map[string]int
	HttpHealthUrl       string
	HealthRemoveArr     []string
	HealthCheckType     string
	HealthCheckTarget   string
}

// 定义隧道结构体
type Tunnel struct {
	Id           int
	Port         int
	ServerIp     string
	Mode         string
	Status       bool
	RunStatus    bool
	Client       *Client
	Ports        string
	Flow         *Flow
	Password     string
	Remark       string
	TargetAddr   string
	NoStore      bool
	LocalPath    string
	StripPre     string
	Target       *Target
	MultiAccount *MultiAccount
	Health
}

/******************************************************************************/

type AjaxResp struct {
	Status int    `json:"status"` // 0: failed, 1: success
	Msg    string `json:"msg"`
}

type AjaxTable[T interface{}] struct {
	Rows  []T `json:"rows"`
	Total int `json:"total"`
}

type AjaxOne[T interface{}] struct {
	Code int `json:"code"` // 0: failed, 1: success
	Data T   `json:"data"`
}
