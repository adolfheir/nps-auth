package cert

import (
	"fmt"
	"testing"
	"time"
)

func TestAdd(t *testing.T) {

	// 生成ca证书
	// GenerateCa()

	// 生成证书请求文件
	csrData, err := GenerateCSR()
	if err != nil {
		panic(err)
	}
	certPEM, ok := csrData["certPEM"].(string)
	if !ok {
		panic("certPEM is not string")
	}

	csrInfo, err := ParseCSR([]byte(certPEM))

	if err != nil {
		panic(err)
	}
	fmt.Printf("csrInfo: %+v\n", csrInfo)

	// 生成证书
	var CertReq = CertData{

		// ChannelId int    `json:"channelId"`
		// Desc      string `json:"desc"` // 备注

		// NpsClientId   string `json:"npsClientId"`
		// NpsClientKey  string `json:"npsClientKey"`
		// NpsTunnelId   string `json:"npsTunnelId"`
		// NpsTunnelPort int    `json:"npsTunnelPort"`

		// MachineId   string `json:"machineId"`
		// ExpiredTime time.Time

		ChannelId: 1,
		Desc:      "test",

		NpsClientId:   "123456",
		NpsClientKey:  "32424",
		NpsTunnelId:   "asd",
		NpsTunnelPort: 1234,

		MachineId:   "123456",
		ExpiredTime: time.Now().AddDate(1, 0, 0),
	}

	certData, err := GenerateCertificate(CertReq)
	if err != nil {
		panic(err)
	}
	fmt.Printf("certData: %+v\n", certData)

	parsedCert, err := ParseCertificate(certData)

	if err != nil {
		panic(err)
	}

	fmt.Printf("parsedCert: %+v\n", parsedCert)

}
