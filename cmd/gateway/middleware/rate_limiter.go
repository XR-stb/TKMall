package middleware

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"TKMall/common/log"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
	"gopkg.in/yaml.v3"
)

// IPRateLimiter 使用令牌桶算法实现的IP限流器
type IPRateLimiter struct {
	ips map[string]*rate.Limiter
	mu  *sync.RWMutex
	// 令牌生产速率(每秒钟产生的令牌数)
	r rate.Limit
	// 令牌桶容量(最大并发数)
	b int
}

// NewIPRateLimiter 创建一个新的IP限流器
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	return &IPRateLimiter{
		ips: make(map[string]*rate.Limiter),
		mu:  &sync.RWMutex{},
		r:   r,
		b:   b,
	}
}

// GetLimiter 获取指定IP的限流器，如果不存在则创建一个新的
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.RLock()
	limiter, exists := i.ips[ip]
	i.mu.RUnlock()

	if !exists {
		i.mu.Lock()
		// 再次检查，避免在获取锁的过程中被其他goroutine创建
		limiter, exists = i.ips[ip]
		if !exists {
			limiter = rate.NewLimiter(i.r, i.b)
			i.ips[ip] = limiter
		}
		i.mu.Unlock()
	}

	return limiter
}

// 清理过期的限流器，避免内存泄漏
func (i *IPRateLimiter) CleanupJob(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		i.mu.Lock()
		for ip := range i.ips {
			delete(i.ips, ip)
		}
		i.mu.Unlock()
		log.Infof("清理IP限流器缓存")
	}
}

// RateLimitConfig 速率限制配置结构
type RateLimitConfig struct {
	Default struct {
		Rate  float64 `yaml:"rate"`
		Burst int     `yaml:"burst"`
	} `yaml:"default"`
	Paths []struct {
		Path      string  `yaml:"path"`
		Rate      float64 `yaml:"rate"`
		Burst     int     `yaml:"burst"`
		UserRate  float64 `yaml:"user_rate"`
		UserBurst int     `yaml:"user_burst"`
	} `yaml:"paths"`
	CleanupInterval int `yaml:"cleanup_interval"`
}

// 存储路径和限流器的映射
var pathLimiters = make(map[string]*IPRateLimiter)

// 存储用户名和限流器的映射
var userLimiters = make(map[string]*IPRateLimiter)

// 默认限流器
var defaultLimiter *IPRateLimiter

// 清理间隔
var cleanupInterval time.Duration

// LoadRateLimitConfig 加载速率限制配置
func LoadRateLimitConfig() error {
	// 读取配置文件
	data, err := os.ReadFile("config/rate_limit.yaml")
	if err != nil {
		return fmt.Errorf("读取速率限制配置失败: %v", err)
	}

	// 解析配置
	var config RateLimitConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("解析速率限制配置失败: %v", err)
	}

	// 设置默认限流器
	defaultLimiter = NewIPRateLimiter(rate.Limit(config.Default.Rate), config.Default.Burst)

	// 设置路径限流器
	for _, p := range config.Paths {
		pathLimiter := NewIPRateLimiter(rate.Limit(p.Rate), p.Burst)
		pathLimiters[p.Path] = pathLimiter

		// 如果有用户限流配置，创建用户限流器
		if p.UserRate > 0 && p.Path == "/login" {
			userLimiters[p.Path] = NewIPRateLimiter(rate.Limit(p.UserRate), p.UserBurst)
		}
	}

	// 设置清理间隔
	cleanupInterval = time.Duration(config.CleanupInterval) * time.Hour
	if cleanupInterval == 0 {
		cleanupInterval = time.Hour // 默认1小时
	}

	log.Infof("加载速率限制配置成功")
	return nil
}

// RateLimiterMiddleware 限流中间件
func RateLimiterMiddleware() gin.HandlerFunc {
	// 确保已加载配置
	if defaultLimiter == nil {
		if err := LoadRateLimitConfig(); err != nil {
			log.Errorf("加载速率限制配置失败: %v，使用默认配置", err)
			defaultLimiter = NewIPRateLimiter(10, 20)
			pathLimiters["/login"] = NewIPRateLimiter(3, 5)
			pathLimiters["/register"] = NewIPRateLimiter(2, 3)
			pathLimiters["/payment"] = NewIPRateLimiter(5, 10)
			pathLimiters["/checkout"] = NewIPRateLimiter(5, 10)
			userLimiters["/login"] = NewIPRateLimiter(1, 5)
			cleanupInterval = time.Hour
		}
	}

	// 启动清理任务
	go func() {
		for path, limiter := range pathLimiters {
			go limiter.CleanupJob(cleanupInterval)
			log.Infof("启动路径 %s 的限流器清理任务", path)
		}
		go defaultLimiter.CleanupJob(cleanupInterval)
		log.Infof("启动默认限流器清理任务")
	}()

	return func(c *gin.Context) {
		var limiter *IPRateLimiter

		// 根据路径选择合适的限流器
		path := c.Request.URL.Path
		if pathLimiter, exists := pathLimiters[path]; exists {
			limiter = pathLimiter
		} else {
			limiter = defaultLimiter
		}

		// 获取客户端IP
		ip := c.ClientIP()
		if ip == "" {
			ip = c.Request.RemoteAddr
		}

		// 获取该IP的限流器
		ipLimiter := limiter.GetLimiter(ip)

		// 尝试获取令牌
		if !ipLimiter.Allow() {
			log.Warnf("IP %s 触发限流保护: %s", ip, path)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "请求过于频繁，请稍后再试",
			})
			c.Abort()
			return
		}

		// 如果是登录接口且配置了用户级别限流，进行用户限流
		if userLimiter, exists := userLimiters[path]; exists && path == "/login" {
			email := c.PostForm("email")
			if email == "" {
				// 尝试从JSON中获取
				var loginData struct {
					Email string `json:"email"`
				}
				if c.ShouldBindJSON(&loginData) == nil && loginData.Email != "" {
					email = loginData.Email
				}
			}

			if email != "" {
				// 获取用户限流器
				userRateLimiter := userLimiter.GetLimiter(email)
				if !userRateLimiter.Allow() {
					log.Warnf("用户 %s 登录失败次数过多", email)
					c.JSON(http.StatusTooManyRequests, gin.H{
						"error": "登录尝试次数过多，请稍后再试",
					})
					c.Abort()
					return
				}
			}
		}

		c.Next()
	}
}
