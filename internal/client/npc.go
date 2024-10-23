package client

import (
	"fmt"
	"nps-auth/configs"
	"nps-auth/pkg/cert"
	"nps-auth/pkg/processManager"
	"os"
	"path"
	"runtime"
	"strings"
)

//./npc -server=175.27.193.51:20102 -vkey=a4ewvo9dspboireu -tls_enable=true

func initNpc() *Npc {

	// 根据系统获取可执行文件路径
	os := runtime.GOOS
	arch := runtime.GOARCH

	fileNames := []string{
		"darwin_amd64_client",
		"linux_arm64_client",
	}
	var matchFile string
	for _, fileName := range fileNames {
		if strings.HasPrefix(fileName, os) && strings.Contains(fileName, arch) {
			matchFile = fileName
		}
	}
	conf := configs.GetConfig()
	fullPath := path.Join(conf.Path, "./data/sidecar/", matchFile, "/npc")

	npc := Npc{
		exePath: fullPath,
	}

	// 如果有cert文件，加载cert文件
	cert, err := LoadCertByFile()

	if err != nil || cert == nil {
		log.Info().Msg("cert is nil")
	} else {
		log.Info().Interface("cert", cert).Msg("cert is exit")
		npc.LoadCert(string(cert))
	}

	return &npc

}

/******************************************************************************
*                                Npc                                   *
******************************************************************************/

type Npc struct {
	exePath  string
	cert     string
	certData *cert.CertData
	process  *processManager.ProcessManager
}

func (npc *Npc) StartNpc() error {
	if npc.process == nil {
		return fmt.Errorf("not authorized")
	}

	stopErr := npc.StopNpc()
	if stopErr != nil {
		log.Error().Err(stopErr).Msg("StopNpc error")
		return stopErr
	}

	err := npc.process.DoStart(true)
	if err != nil {
		return err
	}

	return nil
}

// GetCert 加载证书文件
func (npc *Npc) LoadCert(certStr string) error {

	// 加载证书文件
	certPem := []byte(certStr)
	data, err := cert.ParseCertificate(certPem)
	if err != nil {
		log.Error().Str("cert", certStr).Err(err).Msg("ParseCertificate error")
		return err
	}
	log.Info().Interface("certData", data).Msg("ParseCertificate success")

	// 保存证书信息
	npc.cert = certStr
	npc.certData = data

	// 重置 processManager
	npc.destoryProcess()
	newProcessErr := npc.newProcess()

	if newProcessErr != nil {
		log.Error().Err(newProcessErr).Msg("newProcess error")
		return err
	}

	// 启动进程
	startErr := npc.StartNpc()
	if startErr != nil {
		log.Error().Err(startErr).Msg("StartNpc error")
		return err
	}

	return nil
}

func (npc *Npc) newProcess() error {

	if npc.certData.NpsHost == "" || npc.certData.NpsClientKey == "" {
		return fmt.Errorf("NpsClientKey is not illegal")
	}

	machineid := npc.certData.MachineId
	if npc.certData.MachineId != machineid {
		return fmt.Errorf("machineid is not illegal")
	}

	// 创建 ProcessManager 实例，指定要运行的命令和参数

	// 创建新的实例
	npc.process = processManager.NewProcessManager(
		npc.exePath,
		[]string{
			"-server=" + npc.certData.NpsHost,
			"-vkey=" + npc.certData.NpsClientKey,
			"-tls_enable=" + "true",
			// "-server=175.27.193.51:20102",
			// "-vkey=a4ewvo9dspboireusds",
			// "-tls_enable=true",
		},
		func(output string) processManager.ProcessState {
			if strings.Contains(output, "Validation") {
				npc.cert = ""
				npc.certData = nil
				go npc.destoryProcess()
				return processManager.Stopped
			}

			if strings.Contains(output, "Error") {
				return processManager.Stopped
			}

			if strings.Contains(output, "Successful") {
				return processManager.Running
			}
			return processManager.Starting
		},
	)

	log.Info().Str("path", npc.exePath).Msg("npc process initialized")

	return nil
}

func (npc *Npc) StopNpc() error {
	if npc.process == nil {
		log.Info().Msg("npc process is nil")
	} else {
		err := npc.process.DoStop()
		if err != nil {
			return err
		}
	}

	// 多重保险 kill 残留进程
	killCmdErr := npc.process.KillByCommand()
	if killCmdErr != nil {
		return killCmdErr
	}

	return nil
}

func (npc *Npc) destoryProcess() {
	if npc.process != nil {
		npc.StartNpc()
		npc.process = nil
	}
	log.Info().Msg("npc process destoryed")
}

/******************************************************************************
*                                Index                                   *
******************************************************************************/

// 保存文件到本地
func SaveCertToFile(cert string) error {

	conf := configs.GetConfig()
	fullPath := path.Join(conf.Path, "./data", "/cert.pem")

	certFile, err := os.Create(fullPath)

	if err != nil {
		return err
	}
	defer certFile.Close()

	if _, err := certFile.WriteString(cert); err != nil {
		return err
	}

	return nil
}

func LoadCertByFile() ([]byte, error) {
	conf := configs.GetConfig()
	fullPath := path.Join(conf.Path, "./data", "/cert.pem")

	certData, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}

	return certData, nil
}
