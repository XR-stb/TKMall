package config

import (
	"log"

	"github.com/spf13/viper"
)

func InitConfig(serviceName string) {
	viper.SetConfigName("config")               // 配置文件名 (不带扩展名)
	viper.SetConfigType("yaml")                 // 配置文件类型
	viper.AddConfigPath("./cmd/" + serviceName) // 配置文件路径

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}
}
