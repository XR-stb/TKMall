package middleware

import (
	"net/http"
	"strings"

	"TKMall/cmd/auth/service"
	"TKMall/common/log"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := service.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// 将用户ID存储在上下文中
		c.Set("userID", claims.UserID)
		c.Next()
	}
}

// 检查用户是否在黑名单中
func BlacklistMiddleware(e *casbin.Enforcer) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只对登录接口进行黑名单检查
		if c.Request.URL.Path == "/login" {
			// 从请求中获取用户标识（例如：email）
			email := c.Query("email")
			if email == "" {
				c.Next()
				return
			}

			// 检查用户是否在黑名单中
			ok, err := e.Enforce(email, "/login", "POST")
			if err != nil {
				log.Errorf("黑名单检查失败: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				c.Abort()
				return
			}

			if !ok {
				log.Warnf("黑名单用户尝试登录: %s", email)
				c.JSON(http.StatusForbidden, gin.H{"error": "Account is blocked"})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
