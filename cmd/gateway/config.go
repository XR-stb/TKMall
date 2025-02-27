package main

import (
	"TKMall/common/config"

	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Name string `mapstructure:"name"`
		Port int    `mapstructure:"port"`
	} `mapstructure:"server"`

	Etcd struct {
		Endpoints   []string `mapstructure:"endpoints"`
		DialTimeout int      `mapstructure:"dial_timeout"`
	} `mapstructure:"etcd"`

	Services struct {
		UserService string `mapstructure:"user_service"`
		AuthService string `mapstructure:"auth_service"`
		ProductService string `mapstructure:"product_service"`
		OrderService string `mapstructure:"order_service"`
		PaymentService string `mapstructure:"payment_service"`
		CheckoutService string `mapstructure:"checkout_service"`
		CartService string `mapstructure:"cart_service"`
	} `mapstructure:"services"`
}

func loadConfig() (*Config, error) {
	config.InitConfig("gateway")

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
