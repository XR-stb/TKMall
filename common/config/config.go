package config

import (
	"TKMall/common/log"
	"os"
	"strings"

	"github.com/spf13/viper"
)

func InitConfig(serviceName string) {
	viper.SetConfigName("config")               // 配置文件名 (不带扩展名)
	viper.SetConfigType("yaml")                 // 配置文件类型
	viper.AddConfigPath("./cmd/" + serviceName) // 配置文件路径

	// 添加对环境变量的支持
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	// 支持环境变量覆盖配置
	// 例如，MYSQL_DSN环境变量将覆盖配置中的mysql.dsn
	readEnvVars(serviceName)
}

// 读取特定的环境变量并覆盖配置
func readEnvVars(serviceName string) {
	// 数据库配置
	if dsn := os.Getenv("MYSQL_DSN"); dsn != "" {
		viper.Set("mysql.dsn", dsn)
	}

	// Redis配置
	if redisAddr := os.Getenv("REDIS_ADDR"); redisAddr != "" {
		viper.Set("redis.addr", redisAddr)
	}
	if redisPassword := os.Getenv("REDIS_PASSWORD"); redisPassword != "" {
		viper.Set("redis.password", redisPassword)
	}

	// ETCD配置
	if etcdEndpoints := os.Getenv("ETCD_ENDPOINTS"); etcdEndpoints != "" {
		endpoints := strings.Split(etcdEndpoints, ",")
		viper.Set("etcd.endpoints", endpoints)
	}

	// 服务发现配置 - 使服务名能在k8s中正确注册
	if hostname := os.Getenv("POD_NAME"); hostname != "" && os.Getenv("POD_IP") != "" {
		// 在k8s中，使用POD_IP进行服务注册，而不是localhost
		viper.Set("server.hostname", os.Getenv("POD_IP"))
	}

	// 依赖服务地址配置
	// 用户服务地址
	if addr := os.Getenv("USER_SERVICE_ADDR"); addr != "" {
		viper.Set("user_service.address", addr)
	}

	// 认证服务地址
	if addr := os.Getenv("AUTH_SERVICE_ADDR"); addr != "" {
		viper.Set("auth_service.address", addr)
	}

	// 商品服务地址
	if addr := os.Getenv("PRODUCT_SERVICE_ADDR"); addr != "" {
		viper.Set("product_service.address", addr)
	}

	// 订单服务地址
	if addr := os.Getenv("ORDER_SERVICE_ADDR"); addr != "" {
		viper.Set("order_service.address", addr)
	}

	// 支付服务地址
	if addr := os.Getenv("PAYMENT_SERVICE_ADDR"); addr != "" {
		viper.Set("payment_service.address", addr)
	}

	// 结账服务地址
	if addr := os.Getenv("CHECKOUT_SERVICE_ADDR"); addr != "" {
		viper.Set("checkout_service.address", addr)
	}

	// 购物车服务地址
	if addr := os.Getenv("CART_SERVICE_ADDR"); addr != "" {
		viper.Set("cart_service.address", addr)
	}
}
