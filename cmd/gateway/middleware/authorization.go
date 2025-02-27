package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"TKMall/common/log" // 使用项目的日志包

	"os"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

var enforcer *casbin.Enforcer

// 新增结构体定义
type WhitelistRule struct {
	Path    string   `yaml:"path"`
	Methods []string `yaml:"methods"`
}

var whitelist []WhitelistRule

// 新增路径匹配函数
func matchPath(requestPath, pattern string) bool {
	if pattern == "*" {
		return true
	}
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "/*")
		return strings.HasPrefix(requestPath, prefix)
	}
	return requestPath == pattern
}

// 初始化 Enforcer
func InitEnforcer(e *casbin.Enforcer) {
	enforcer = e
	// 启动定期重新加载
	go autoReload()
}

// 定期重新加载策略
func autoReload() {
	ticker := time.NewTicker(5 * time.Second) // 每5秒重新加载一次
	for range ticker.C {
		err := enforcer.LoadPolicy()
		if err != nil {
			log.Errorf("Failed to reload policy: %v", err)
		} else {
			log.Infof("Policy reloaded successfully")
		}
	}
}

// 修改中间件逻辑
func AuthorizationMiddleware(e *casbin.Enforcer) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 白名单检查
		for _, rule := range whitelist {
			if matchPath(c.Request.URL.Path, rule.Path) {
				for _, method := range rule.Methods {
					if strings.EqualFold(method, c.Request.Method) {
						log.Infof("跳过白名单接口: %s %s", c.Request.Method, rule.Path)
						c.Next()
						return
					}
				}
			}
		}

		log.Infof("正在处理请求: %s %s", c.Request.Method, c.Request.URL.Path)

		// 只跳过注册接口的权限验证
		if c.Request.URL.Path == "/register" {
			log.Infof("跳过注册接口的权限验证")
			c.Next()
			return
		}

		// 对于登录接口，使用 email 作为主体进行验证
		if c.Request.URL.Path == "/login" {
			c.Next()
			return
		}

		// 其他接口的权限验证
		userID, exists := c.Get("userID")
		log.Infof("其他接口验证，userID=%v, exists=%v", userID, exists)
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		// 获取请求路径和方法
		obj := c.Request.URL.Path
		act := c.Request.Method

		// 检查权限
		ok, err := e.Enforce(fmt.Sprint(userID), obj, act)
		if err != nil {
			log.Errorf("权限检查失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking permissions"})
			c.Abort()
			return
		}

		if !ok {
			log.Warnf("用户 %v 访问 %s %s 被拒绝", userID, act, obj)
			c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
			c.Abort()
			return
		}

		log.Infof("用户 %v 访问 %s %s 已授权", userID, act, obj)
		c.Next()
	}
}

// 更新用户角色
func UpdateUserRole(e *casbin.Enforcer, userID string, role string) error {
	// 移除用户的所有现有角色
	e.RemoveFilteredGroupingPolicy(0, userID)

	// 添加新角色
	_, err := e.AddGroupingPolicy(userID, role)
	if err != nil {
		return err
	}

	// 加载更新后的策略
	return e.LoadPolicy()
}

// 禁用用户
func DisableUser(e *casbin.Enforcer, userID string) error {
	// 移除用户的所有角色
	_, err := e.RemoveFilteredGroupingPolicy(0, userID)
	if err != nil {
		return err
	}

	// 加载更新后的策略
	return e.LoadPolicy()
}

// 将用户加入黑名单
func BlockUser(e *casbin.Enforcer, email string) error {
	// 添加黑名单策略
	_, err := e.AddGroupingPolicy(email, "blocked_user")
	if err != nil {
		return err
	}

	// 直接调用 LoadPolicy 即可
	return e.LoadPolicy()
}

// 将用户从黑名单移除
func UnblockUser(e *casbin.Enforcer, email string) error {
	// 移除黑名单策略
	_, err := e.RemoveFilteredGroupingPolicy(0, email, "blocked_user")
	if err != nil {
		return err
	}

	// 加载更新后的策略
	return e.LoadPolicy()
}

// 批量添加用户到黑名单
func BlockUsers(e *casbin.Enforcer, emails []string) error {
	for _, email := range emails {
		_, err := e.AddGroupingPolicy(email, "blocked_user")
		if err != nil {
			return fmt.Errorf("failed to block user %s: %v", email, err)
		}
	}

	// 加载更新后的策略
	return e.LoadPolicy()
}

// 批量从黑名单移除用户
func UnblockUsers(e *casbin.Enforcer, emails []string) error {
	for _, email := range emails {
		_, err := e.RemoveFilteredGroupingPolicy(0, email, "blocked_user")
		if err != nil {
			return fmt.Errorf("failed to unblock user %s: %v", email, err)
		}
	}

	// 加载更新后的策略
	return e.LoadPolicy()
}

// 获取所有黑名单用户
func GetBlockedUsers(e *casbin.Enforcer) []string {
	policies, _ := e.GetFilteredGroupingPolicy(1, "blocked_user")
	users := make([]string, len(policies))
	for i, policy := range policies {
		users[i] = policy[0]
	}
	return users
}

// 重新加载策略
func ReloadPolicy(e *casbin.Enforcer) error {
	return e.LoadPolicy()
}

func LoadWhitelistConfig() error {
	data, err := os.ReadFile("config/security.yaml")
	if err != nil {
		return fmt.Errorf("读取白名单配置失败: %v", err)
	}

	config := struct {
		Whitelist []WhitelistRule `yaml:"whitelist"`
	}{}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("解析白名单配置失败: %v", err)
	}

	whitelist = config.Whitelist
	return nil
}
