package cert

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"sync"
)

var (
	MachineID string
	once      sync.Once
)

func GetMachineID() string {
	once.Do(func() {
		MachineID = genMachineId()
	})
	return MachineID
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
