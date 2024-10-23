package configs

import (
	"fmt"
	"os"
	"sync"

	"github.com/spf13/viper"
)

type Config struct {
	Http   Http
	Logger Logger
	DB     DB
}

var (
	EnvConfig *Config
	once      sync.Once
)

func loadConfig() *Config {

	path, err := os.Getwd() // get curent path
	if err != nil {
		panic(err)
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path + "/configs") // path to look for the config file in

	if err := viper.ReadInConfig(); err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		panic(err)
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
