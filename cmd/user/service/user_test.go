package service

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// 设置测试路由
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// 创建用户API路由组
	userGroup := r.Group("/user")
	{
		userGroup.POST("/register", func(c *gin.Context) {
			var req struct {
				Username string `json:"username"`
				Password string `json:"password"`
				Email    string `json:"email"`
				Phone    string `json:"phone"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			// 参数校验
			if req.Username == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "用户名不能为空"})
				return
			}
			if req.Password == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "密码不能为空"})
				return
			}
			if req.Email == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "邮箱不能为空"})
				return
			}

			// 模拟用户已存在的情况
			if req.Username == "existing_user" {
				c.JSON(http.StatusConflict, gin.H{"error": "用户名已存在"})
				return
			}

			// 模拟邮箱已存在的情况
			if req.Email == "existing@example.com" {
				c.JSON(http.StatusConflict, gin.H{"error": "邮箱已被注册"})
				return
			}

			// 模拟服务器错误
			if req.Username == "error_user" {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
				return
			}

			// 成功注册
			c.JSON(http.StatusOK, gin.H{
				"user_id": 12345,
				"message": "注册成功",
			})
		})

		userGroup.POST("/login", func(c *gin.Context) {
			var req struct {
				Username string `json:"username"`
				Password string `json:"password"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			// 参数校验
			if req.Username == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "用户名不能为空"})
				return
			}
			if req.Password == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "密码不能为空"})
				return
			}

			// 模拟用户不存在
			if req.Username == "nonexistent_user" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
				return
			}

			// 模拟密码错误
			if req.Username == "valid_user" && req.Password != "correct_password" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "密码错误"})
				return
			}

			// 模拟服务器错误
			if req.Username == "error_user" {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
				return
			}

			// 成功登录
			c.JSON(http.StatusOK, gin.H{
				"user_id":  12345,
				"username": req.Username,
				"token":    "mock_jwt_token",
				"message":  "登录成功",
			})
		})

		userGroup.GET("/profile", func(c *gin.Context) {
			userID := c.Query("user_id")
			if userID == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID不能为空"})
				return
			}

			// 模拟用户不存在
			if userID == "999999" {
				c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
				return
			}

			// 模拟服务器错误
			if userID == "888888" {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
				return
			}

			// 返回用户资料
			c.JSON(http.StatusOK, gin.H{
				"user_id":    userID,
				"username":   "test_user",
				"email":      "test@example.com",
				"phone":      "13812345678",
				"created_at": "2023-05-01T12:00:00Z",
				"last_login": "2023-05-20T15:30:00Z",
			})
		})
	}

	return r
}

// 测试用户注册功能
func TestRegisterUser(t *testing.T) {
	router := setupTestRouter()

	t.Run("参数验证", func(t *testing.T) {
		// 测试空用户名
		jsonData := `{
			"username": "",
			"password": "password123",
			"email": "test@example.com",
			"phone": "13812345678"
		}`
		req, _ := http.NewRequest("POST", "/user/register", bytes.NewBufferString(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "空用户名应返回400")
		assert.Contains(t, w.Body.String(), "用户名不能为空", "响应应包含错误信息")

		// 测试空密码
		jsonData = `{
			"username": "testuser",
			"password": "",
			"email": "test@example.com",
			"phone": "13812345678"
		}`
		req, _ = http.NewRequest("POST", "/user/register", bytes.NewBufferString(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "空密码应返回400")
		assert.Contains(t, w.Body.String(), "密码不能为空", "响应应包含错误信息")

		// 测试空邮箱
		jsonData = `{
			"username": "testuser",
			"password": "password123",
			"email": "",
			"phone": "13812345678"
		}`
		req, _ = http.NewRequest("POST", "/user/register", bytes.NewBufferString(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "空邮箱应返回400")
		assert.Contains(t, w.Body.String(), "邮箱不能为空", "响应应包含错误信息")
	})

	t.Run("用户已存在", func(t *testing.T) {
		jsonData := `{
			"username": "existing_user",
			"password": "password123",
			"email": "test@example.com",
			"phone": "13812345678"
		}`
		req, _ := http.NewRequest("POST", "/user/register", bytes.NewBufferString(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code, "用户已存在应返回409")
		assert.Contains(t, w.Body.String(), "用户名已存在", "响应应包含错误信息")
	})

	t.Run("邮箱已存在", func(t *testing.T) {
		jsonData := `{
			"username": "newuser",
			"password": "password123",
			"email": "existing@example.com",
			"phone": "13812345678"
		}`
		req, _ := http.NewRequest("POST", "/user/register", bytes.NewBufferString(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code, "邮箱已存在应返回409")
		assert.Contains(t, w.Body.String(), "邮箱已被注册", "响应应包含错误信息")
	})

	t.Run("服务器错误", func(t *testing.T) {
		jsonData := `{
			"username": "error_user",
			"password": "password123",
			"email": "error@example.com",
			"phone": "13812345678"
		}`
		req, _ := http.NewRequest("POST", "/user/register", bytes.NewBufferString(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code, "服务器错误应返回500")
		assert.Contains(t, w.Body.String(), "服务器内部错误", "响应应包含错误信息")
	})

	t.Run("成功注册", func(t *testing.T) {
		jsonData := `{
			"username": "newuser",
			"password": "password123",
			"email": "new@example.com",
			"phone": "13812345678"
		}`
		req, _ := http.NewRequest("POST", "/user/register", bytes.NewBufferString(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "成功注册应返回200")

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, "应返回有效的JSON")
		assert.Equal(t, float64(12345), response["user_id"], "应返回用户ID")
		assert.Contains(t, response["message"], "注册成功", "响应应包含成功信息")
	})
}

// 测试用户登录功能
func TestLoginUser(t *testing.T) {
	router := setupTestRouter()

	t.Run("参数验证", func(t *testing.T) {
		// 测试空用户名
		jsonData := `{
			"username": "",
			"password": "password123"
		}`
		req, _ := http.NewRequest("POST", "/user/login", bytes.NewBufferString(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "空用户名应返回400")
		assert.Contains(t, w.Body.String(), "用户名不能为空", "响应应包含错误信息")

		// 测试空密码
		jsonData = `{
			"username": "testuser",
			"password": ""
		}`
		req, _ = http.NewRequest("POST", "/user/login", bytes.NewBufferString(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "空密码应返回400")
		assert.Contains(t, w.Body.String(), "密码不能为空", "响应应包含错误信息")
	})

	t.Run("用户不存在", func(t *testing.T) {
		jsonData := `{
			"username": "nonexistent_user",
			"password": "password123"
		}`
		req, _ := http.NewRequest("POST", "/user/login", bytes.NewBufferString(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code, "用户不存在应返回401")
		assert.Contains(t, w.Body.String(), "用户不存在", "响应应包含错误信息")
	})

	t.Run("密码错误", func(t *testing.T) {
		jsonData := `{
			"username": "valid_user",
			"password": "wrong_password"
		}`
		req, _ := http.NewRequest("POST", "/user/login", bytes.NewBufferString(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code, "密码错误应返回401")
		assert.Contains(t, w.Body.String(), "密码错误", "响应应包含错误信息")
	})

	t.Run("服务器错误", func(t *testing.T) {
		jsonData := `{
			"username": "error_user",
			"password": "password123"
		}`
		req, _ := http.NewRequest("POST", "/user/login", bytes.NewBufferString(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code, "服务器错误应返回500")
		assert.Contains(t, w.Body.String(), "服务器内部错误", "响应应包含错误信息")
	})

	t.Run("成功登录", func(t *testing.T) {
		jsonData := `{
			"username": "valid_user",
			"password": "correct_password"
		}`
		req, _ := http.NewRequest("POST", "/user/login", bytes.NewBufferString(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "成功登录应返回200")

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, "应返回有效的JSON")
		assert.Equal(t, float64(12345), response["user_id"], "应返回用户ID")
		assert.Equal(t, "valid_user", response["username"], "应返回用户名")
		assert.Equal(t, "mock_jwt_token", response["token"], "应返回令牌")
		assert.Contains(t, response["message"], "登录成功", "响应应包含成功信息")
	})
}

// 测试获取用户资料功能
func TestGetUserProfile(t *testing.T) {
	router := setupTestRouter()

	t.Run("参数验证", func(t *testing.T) {
		// 测试空用户ID
		req, _ := http.NewRequest("GET", "/user/profile", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "空用户ID应返回400")
		assert.Contains(t, w.Body.String(), "用户ID不能为空", "响应应包含错误信息")
	})

	t.Run("用户不存在", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/user/profile?user_id=999999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code, "用户不存在应返回404")
		assert.Contains(t, w.Body.String(), "用户不存在", "响应应包含错误信息")
	})

	t.Run("服务器错误", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/user/profile?user_id=888888", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code, "服务器错误应返回500")
		assert.Contains(t, w.Body.String(), "服务器内部错误", "响应应包含错误信息")
	})

	t.Run("成功获取资料", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/user/profile?user_id=12345", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "成功获取资料应返回200")

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, "应返回有效的JSON")
		assert.Equal(t, "12345", response["user_id"], "应返回正确的用户ID")
		assert.Equal(t, "test_user", response["username"], "应返回正确的用户名")
		assert.Equal(t, "test@example.com", response["email"], "应返回正确的邮箱")
		assert.Equal(t, "13812345678", response["phone"], "应返回正确的手机号")
		assert.Contains(t, response["created_at"], "2023-05-01", "应返回注册时间")
	})
}
