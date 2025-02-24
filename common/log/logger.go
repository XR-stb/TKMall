package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v2"
)

var (
	instance *logrus.Logger
	once     sync.Once
)

type Config struct {
	LogDir     string `yaml:"log_dir"`     // 日志存储目录
	Format     string `yaml:"format"`      // 日志格式：json/text
	MaxSize    int    `yaml:"max_size"`    // 单个日志文件最大大小(MB)
	MaxBackups int    `yaml:"max_backups"` // 保留旧日志最大数量
	MaxAge     int    `yaml:"max_age"`     // 保留旧日志最大天数
	Compress   bool   `yaml:"compress"`    // 是否压缩旧日志
	Level      string `yaml:"level"`       // 日志级别
}

// Init 初始化日志，serviceName 由调用方传入
func Init(configPath string, serviceName string) {
	once.Do(func() {
		// 读取配置文件
		cfg := loadConfig(configPath)

		// 初始化日志实例
		instance = logrus.New()

		// 启用调用者信息记录
		instance.SetReportCaller(true)

		// 自定义调用者信息格式化
		prettyfier := func(f *runtime.Frame) (string, string) {
			// 获取完整的调用栈
			pcs := make([]uintptr, 10)
			n := runtime.Callers(9, pcs)
			if n == 0 {
				return "unknown", "unknown:0"
			}

			frames := runtime.CallersFrames(pcs[:n])
			// 跳过日志包装层
			for {
				frame, more := frames.Next()
				if !strings.Contains(frame.Function, "TKMall/common/log") {
					// 获取函数名，去掉项目前缀
					funcName := strings.TrimPrefix(frame.Function, "TKMall/")

					// 获取文件路径，去掉项目根路径
					filePath := frame.File
					if idx := strings.Index(filePath, "TKMall/"); idx != -1 {
						filePath = filePath[idx+7:] // 7 是 "TKMall/" 的长度
					}

					return funcName, fmt.Sprintf("%s:%d", filePath, frame.Line)
				}
				if !more {
					break
				}
			}
			return f.Function, fmt.Sprintf("%s:%d", filepath.Base(f.File), f.Line)
		}

		// 配置日志格式
		switch cfg.Format {
		case "json":
			instance.SetFormatter(&logrus.JSONFormatter{
				TimestampFormat:  "2006-01-02 15:04:05",
				CallerPrettyfier: prettyfier,
			})
		default:
			instance.SetFormatter(&logrus.TextFormatter{
				FullTimestamp:    true,
				TimestampFormat:  "2006-01-02 15:04:05",
				CallerPrettyfier: prettyfier,
			})
		}

		// 配置日志输出
		logPath := filepath.Join(cfg.LogDir, serviceName) // 使用传入的服务名
		os.MkdirAll(logPath, 0755)

		lumberjackLogger := &lumberjack.Logger{
			Filename:   filepath.Join(logPath, "app.log"),
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}

		// 同时输出到文件和控制台
		instance.SetOutput(io.MultiWriter(os.Stdout, lumberjackLogger))

		// 设置日志级别
		if level, err := logrus.ParseLevel(cfg.Level); err == nil {
			instance.SetLevel(level)
		} else {
			instance.SetLevel(logrus.InfoLevel) // 默认级别
		}
	})
}

func loadConfig(path string) *Config {
	// 读取配置文件内容
	configFile, err := os.Open(path)
	if err != nil {
		logrus.Fatalf("无法打开日志配置文件: %v", err)
	}
	defer configFile.Close()

	// 解析YAML配置
	var cfg Config
	decoder := yaml.NewDecoder(configFile)
	if err := decoder.Decode(&cfg); err != nil {
		logrus.Fatalf("解析日志配置失败: %v", err)
	}

	// 设置默认值
	if cfg.LogDir == "" {
		cfg.LogDir = "./logs"
	}
	if cfg.Format == "" {
		cfg.Format = "json"
	}
	if cfg.MaxSize == 0 {
		cfg.MaxSize = 100 // 默认100MB
	}
	if cfg.MaxBackups == 0 {
		cfg.MaxBackups = 7 // 默认保留7天
	}
	if cfg.MaxAge == 0 {
		cfg.MaxAge = 30 // 默认保留30天
	}

	return &cfg
}

// 恢复使用直接的实例方法调用
func Debug(args ...interface{}) {
	instance.Debug(args...)
}

func Info(args ...interface{}) {
	instance.Info(args...)
}

func Warn(args ...interface{}) {
	instance.Warn(args...)
}

func Error(args ...interface{}) {
	instance.Error(args...)
}

func Fatal(args ...interface{}) {
	instance.Fatal(args...)
}

// 带格式化的方法
func Debugf(format string, args ...interface{}) {
	instance.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	instance.Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	instance.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	instance.Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	instance.Fatalf(format, args...)
}

// WithFields 也需要修改
func WithFields(fields map[string]interface{}) *logrus.Entry {
	return instance.WithFields(logrus.Fields(fields))
}
