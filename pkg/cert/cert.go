package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"embed"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/denisbrodbeck/machineid"
)

/******************************************************************************
*                                   root ca                                   *
******************************************************************************/

//go:embed ca_cert.pem
//go:embed ca_key.pem
var fsCA embed.FS

func GenerateCa() {
	// 生成 RSA 私钥
	caPriv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Println("生成 CA 私钥失败:", err)
		return
	}

	// 生成 CA 证书
	caTemplate := x509.Certificate{
		SerialNumber: big.NewInt(38217062821277),
		Subject: pkix.Name{
			Organization: []string{"CA IHOUQI"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(100, 0, 0), // 设置证书有效期为100年

		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IsCA:                  true, // 标记为 CA 证书
		BasicConstraintsValid: true, // 确保基本约束有效
	}

	// 创建 CA 证书
	caCertDER, err := x509.CreateCertificate(rand.Reader, &caTemplate, &caTemplate, &caPriv.PublicKey, caPriv)
	if err != nil {
		fmt.Println("生成 CA 证书失败:", err)
		return
	}

	// 保存 CA 私钥
	caPrivFile, err := os.Create("ca_key.pem")
	if err != nil {
		fmt.Println("创建 CA 私钥文件失败:", err)
		return
	}
	defer caPrivFile.Close()

	if err := pem.Encode(caPrivFile, &pem.Block{Type: "PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(caPriv)}); err != nil {
		fmt.Println("保存 CA 私钥失败:", err)
		return
	}

	// 保存 CA 证书
	caCertFile, err := os.Create("ca_cert.pem")
	if err != nil {
		fmt.Println("创建 CA 证书文件失败:", err)
		return
	}
	defer caCertFile.Close()

	if err := pem.Encode(caCertFile, &pem.Block{Type: "CERTIFICATE", Bytes: caCertDER}); err != nil {
		fmt.Println("保存 CA 证书失败:", err)
		return
	}

}

func loadCAKey() (*rsa.PrivateKey, error) {
	caKeyData, err := fsCA.ReadFile("ca_key.pem")
	if err != nil {
		return nil, err
	}

	caKeyBlock, _ := pem.Decode(caKeyData)
	if caKeyBlock == nil || caKeyBlock.Type != "PRIVATE KEY" {
		return nil, fmt.Errorf("无法解析 CA 私钥")
	}

	caPriv, err := x509.ParsePKCS1PrivateKey(caKeyBlock.Bytes)
	if err != nil {
		return nil, err
	}

	return caPriv, nil
}

func loadCACert() (*x509.Certificate, error) {
	caCertData, err := fsCA.ReadFile("ca_cert.pem")
	if err != nil {
		return nil, err
	}

	caCertBlock, _ := pem.Decode(caCertData)
	if caCertBlock == nil || caCertBlock.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("无法解析 CA 证书")
	}

	caCert, err := x509.ParseCertificate(caCertBlock.Bytes)
	if err != nil {
		return nil, err
	}

	return caCert, nil
}

/******************************************************************************
*                                生成/解析 证书                                 *
******************************************************************************/

// OID（Object Identifier）1.3.6.1.4.1.12345 解析：
//   - 1.3：表示ISO（国际标准化组织）和ITU-T（国际电信联盟）共同制定的OID树的部分。
//   - 6：表示ISO成员国中的美国。
//   - 1：表示ANSI（美国国家标准协会）。
//   - 4.1：表示私人使用的OID。
//   - 12345：表示私人使用的OID 这里代表ihouqi。
var RootOID = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 12345} // 假设根OID为1.3.6.1.4.1.12345
var CsrDataOID = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 12345, 1}
var CertDataOID = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 12345, 2}

type CsrData struct {
	MachineId string
	Time      time.Time
}

// 生成证书请求文件（CSR）并包含机器码
func GenerateCSR() ([]byte, error) {
	// 这里的私钥可以任意生成 只做签名用
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	machineId, err := machineid.ID()
	if err != nil {
		return nil, err
	}

	csrDataHash, err := asn1.Marshal(CsrData{MachineId: machineId, Time: time.Now()})
	if err != nil {
		return nil, err
	}

	csrTemplate := x509.CertificateRequest{
		Subject: pkix.Name{
			Organization: []string{"ihouqi"},
			CommonName:   "https://ihouqi.cn/",
		},
		ExtraExtensions: []pkix.Extension{
			{
				Id:       CsrDataOID,
				Value:    csrDataHash,
				Critical: false,
			},
		},
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, privateKey)
	if err != nil {
		return nil, err
	}

	csrPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes})
	return csrPEM, nil
}

// 解析证书请求文件并提取机器码
func ParseCSR(csrPEM []byte) (*CsrData, error) {
	csrBlock, _ := pem.Decode(csrPEM)
	if csrBlock == nil || csrBlock.Type != "CERTIFICATE REQUEST" {
		return nil, fmt.Errorf("无法解析证书请求")
	}

	csr, err := x509.ParseCertificateRequest(csrBlock.Bytes)
	if err != nil {
		return nil, err
	}

	for _, ext := range csr.Extensions {
		if ext.Id.Equal(CsrDataOID) {
			var csrData CsrData
			if _, err := asn1.Unmarshal(ext.Value, &csrData); err != nil {
				return nil, err
			}
			return &csrData, nil
		}
	}

	return nil, fmt.Errorf("证书请求中未包含机器码")
}

// 生成证书
type CertData struct {
	ChannelId int    `json:"channelId"`
	Desc      string `json:"desc"` // 备注

	NpsHost       string `json:"npsHost"`
	NpsClientId   string `json:"npsClientId"`
	NpsClientKey  string `json:"npsClientKey"`
	NpsTunnelId   string `json:"npsTunnelId"`
	NpsTunnelPort int    `json:"npsTunnelPort"`

	MachineId   string `json:"machineId"`
	ExpiredTime time.Time
}

func GenerateCertificate(CertReq CertData) ([]byte, error) {

	// 获取ca
	caKey, err := loadCAKey()
	if err != nil {
		return nil, err
	}

	caCert, err := loadCACert()
	if err != nil {
		return nil, err
	}

	// 生成证书
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	certDataHash, err := asn1.Marshal(CertReq)
	if err != nil {
		return nil, err
	}

	certTemplate := x509.Certificate{
		SerialNumber: big.NewInt(382170628212771),
		Subject: pkix.Name{
			Organization: []string{"ihouqi"},
			CommonName:   "https://ihouqi.cn/",
		},
		NotBefore:             time.Now(),                  // 设置证书生效时间为当前时间
		NotAfter:              time.Now().AddDate(1, 0, 0), // 设置证书有效期为一年
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,  // 确保基本约束有效
		IsCA:                  false, // 这是一个非CA证书
		ExtraExtensions: []pkix.Extension{
			{
				Id:       CertDataOID,
				Value:    certDataHash,
				Critical: false,
			},
		},
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &certTemplate, caCert, &privateKey.PublicKey, caKey)
	if err != nil {
		return nil, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	return certPEM, nil
}

// 解析证书信息 校验ca 并返回CertData
func ParseCertificate(certPEM []byte) (*CertData, error) {
	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil || certBlock.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("无法解析证书")
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, err
	}

	caCert, err := loadCACert()
	if err != nil {
		return nil, err
	}

	roots := x509.NewCertPool()
	roots.AddCert(caCert)

	opts := x509.VerifyOptions{
		Roots: roots,
	}

	if _, err := cert.Verify(opts); err != nil {
		return nil, err
	}

	for _, ext := range cert.Extensions {
		if ext.Id.Equal(CertDataOID) {
			var certData CertData
			if _, err := asn1.Unmarshal(ext.Value, &certData); err != nil {
				return nil, err
			}
			return &certData, nil
		}
	}

	return nil, fmt.Errorf("证书中未包含信息")
}

/******************************************************************************
*                                Index                                   *
******************************************************************************/

// 保存文件到本地
func SaveCert(cert string) error {
	certFile, err := os.Create("cert.pem")
	if err != nil {
		return err
	}
	defer certFile.Close()

	if _, err := certFile.WriteString(cert); err != nil {
		return err
	}

	return nil
}

func LoadCert() ([]byte, error) {
	certData, err := os.ReadFile("cert.pem")
	if err != nil {
		return nil, err
	}

	return certData, nil
}
