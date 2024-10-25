package configs

import (
	"fmt"
	"os"
	"path"
	"sync"

	"github.com/spf13/viper"
)

type Config struct {
	Path   string
	Http   Http
	Logger Logger
	DB     DB
	Nps    Nps
}

var (
	EnvConfig *Config
	once      sync.Once
)

func loadConfig() *Config {

	wd, err := os.Getwd() // get curent path
	if err != nil {
		panic(err)
	}

	// 设置默认值到 viper
	defaultConfig := &Config{
		Path: wd,
		Http: Http{
			ClientAddr: ":20108",
			ServerAddr: ":20109",
		},
		Logger: Logger{
			Level:  "info",
			Output: "file",
		},
		DB: DB{
			DSN: "./data/nps.sqlite3",
		},
		Nps: Nps{
			ApiHost:    "http://175.27.193.51:20100",
			ApiKey:     "ihouqi2022",
			BridgeHost: "175.27.193.51:20102",
			ClientPort: "32301",
		},
	}
	viper.SetDefault("path", defaultConfig.Path)
	viper.SetDefault("http.clientAddr", defaultConfig.Http.ClientAddr)
	viper.SetDefault("http.serverAddr", defaultConfig.Http.ServerAddr)
	viper.SetDefault("logger.level", defaultConfig.Logger.Level)
	viper.SetDefault("logger.output", defaultConfig.Logger.Output)
	viper.SetDefault("db.dsn", defaultConfig.DB.DSN)
	viper.SetDefault("db.dsn", defaultConfig.DB.DSN)
	viper.SetDefault("nps.apiHost", defaultConfig.Nps.ApiHost)
	viper.SetDefault("nps.apiKey", defaultConfig.Nps.ApiKey)
	viper.SetDefault("nps.bridgeHost", defaultConfig.Nps.BridgeHost)
	viper.SetDefault("nps.clientPort", defaultConfig.Nps.ClientPort)

	// 读取配置文件
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/nps-auth")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil { // Handle errors reading the config file
		fmt.Println("config.yaml is not find, use default")
	} else {
		fmt.Printf("Used configuration file is: %s\n", viper.ConfigFileUsed())
	}

	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	// 检查数据文件夹是否存在
	var folderPath = path.Join(config.Path, "./data/")
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		// 如果文件夹不存在，则创建文件夹
		err := os.MkdirAll(folderPath, os.ModePerm)
		if err != nil {
			panic(fmt.Errorf("create data path err: %w", err))
		}
	}

	return config
}

func GetConfig() *Config {

	once.Do(func() {
		EnvConfig = loadConfig()
		fmt.Printf("init conf:  %+v \n", EnvConfig)
	})

	return EnvConfig
}
