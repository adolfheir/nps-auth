package cert

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"nps-auth/configs"
	"os"
	"path"
	"strings"
	"sync"
)

var (
	MachineID string
	once      sync.Once
	idFile    = "./machine_id.tmp" // 定义保存 MachineID 的文件路径
)

func GetMachineID() string {
	once.Do(func() {
		MachineID = loadOrCreateMachineID()
	})
	return MachineID
}

// 尝试从文件中加载 MachineID，如果文件不存在，则生成并保存
func loadOrCreateMachineID() string {

	conf := configs.GetConfig()
	fullPath := path.Join(conf.Path, "./data", idFile)

	// 检查文件是否存在
	if _, err := os.Stat(fullPath); err == nil {
		// 文件存在，读取 MachineID
		idBytes, err := os.ReadFile(fullPath)
		if err != nil {
			panic(fmt.Sprintf("无法读取机器ID文件: %v", err))
		}
		return string(idBytes)
	}

	// 文件不存在，生成新的 MachineID 并保存
	newID := genMachineId()

	// 将生成的 MachineID 写入文件
	err := os.WriteFile(fullPath, []byte(newID), 0644)
	if err != nil {
		panic(fmt.Sprintf("无法保存机器ID到文件: %v", err))
	}

	return newID
}

func genMachineId() string {
	// 获取所有网卡信息
	interfaces, err := net.Interfaces()
	if err != nil {
		panic(fmt.Sprintf("无法获取网卡信息: %v", err))
	}

	var machineCode []string

	for _, iface := range interfaces {
		// 获取网卡的 MAC 地址
		mac := iface.HardwareAddr
		if mac != nil {
			machineCode = append(machineCode, mac.String())
		}
	}

	// 组合机器码
	result := strings.Join(machineCode, "-")

	// 计算 SHA-256 哈希
	hash := sha256.New()
	hash.Write([]byte(result))
	hashedValue := hash.Sum(nil)

	// 将哈希值转为十六进制字符串
	hashString := hex.EncodeToString(hashedValue)

	return hashString
}
