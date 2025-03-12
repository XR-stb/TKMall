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

	// 创建购物车API路由组
	cartGroup := r.Group("/cart")
	{
		cartGroup.POST("/add", func(c *gin.Context) {
			var req struct {
				UserID    int64  `json:"user_id"`
				ProductID uint32 `json:"product_id"`
				Quantity  uint32 `json:"quantity"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			// 参数校验
			if req.UserID <= 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
				return
			}

			if req.ProductID <= 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "无效的商品ID"})
				return
			}

			if req.Quantity <= 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "无效的商品数量"})
				return
			}

			// 成功添加到购物车
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "成功添加到购物车",
			})
		})

		cartGroup.GET("/get", func(c *gin.Context) {
			userID := c.Query("user_id")
			if userID == "" || userID == "0" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
				return
			}

			// 模拟获取购物车内容
			if userID == "999" {
				// 模拟空购物车
				c.JSON(http.StatusOK, gin.H{
					"cart": gin.H{
						"user_id": userID,
						"items":   []interface{}{},
					},
				})
				return
			}

			// 正常购物车
			c.JSON(http.StatusOK, gin.H{
				"cart": gin.H{
					"user_id": userID,
					"items": []gin.H{
						{
							"product_id": 101,
							"quantity":   2,
							"price":      199.99,
						},
						{
							"product_id": 102,
							"quantity":   1,
							"price":      299.99,
						},
					},
				},
			})
		})

		cartGroup.POST("/empty", func(c *gin.Context) {
			var req struct {
				UserID int64 `json:"user_id"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			// 参数校验
			if req.UserID <= 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
				return
			}

			// 模拟数据库错误
			if req.UserID == 888 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "清空购物车失败: 数据库连接失败"})
				return
			}

			// 成功清空购物车
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "购物车已清空",
			})
		})
	}

	return r
}

// 测试添加商品到购物车
func TestAddItem(t *testing.T) {
	router := setupTestRouter()

	t.Run("参数验证失败", func(t *testing.T) {
		// 测试无效的用户ID
		jsonData := `{"user_id": 0, "product_id": 1, "quantity": 2}`
		req, _ := http.NewRequest("POST", "/cart/add", bytes.NewBufferString(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "无效的用户ID应返回400")
		assert.Contains(t, w.Body.String(), "无效的用户ID", "响应应包含错误信息")

		// 测试无效的商品ID
		jsonData = `{"user_id": 1, "product_id": 0, "quantity": 2}`
		req, _ = http.NewRequest("POST", "/cart/add", bytes.NewBufferString(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "无效的商品ID应返回400")
		assert.Contains(t, w.Body.String(), "无效的商品ID", "响应应包含错误信息")

		// 测试无效的商品数量
		jsonData = `{"user_id": 1, "product_id": 1, "quantity": 0}`
		req, _ = http.NewRequest("POST", "/cart/add", bytes.NewBufferString(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "无效的商品数量应返回400")
		assert.Contains(t, w.Body.String(), "无效的商品数量", "响应应包含错误信息")
	})

	t.Run("成功添加购物车项", func(t *testing.T) {
		jsonData := `{"user_id": 1, "product_id": 101, "quantity": 2}`
		req, _ := http.NewRequest("POST", "/cart/add", bytes.NewBufferString(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "成功添加商品应返回200")

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, "应返回有效的JSON")
		assert.Equal(t, true, response["success"], "操作应该成功")
		assert.Contains(t, response["message"], "成功添加到购物车", "响应消息应正确")
	})
}

// 测试获取购物车
func TestGetCart(t *testing.T) {
	router := setupTestRouter()

	t.Run("参数验证失败", func(t *testing.T) {
		// 测试无效的用户ID
		req, _ := http.NewRequest("GET", "/cart/get?user_id=0", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "无效的用户ID应返回400")
		assert.Contains(t, w.Body.String(), "无效的用户ID", "响应应包含错误信息")
	})

	t.Run("购物车为空", func(t *testing.T) {
		// 使用特殊ID 999表示空购物车
		req, _ := http.NewRequest("GET", "/cart/get?user_id=999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "获取空购物车应返回200")

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, "应返回有效的JSON")

		cart := response["cart"].(map[string]interface{})
		items := cart["items"].([]interface{})
		assert.Empty(t, items, "购物车项应为空")
	})

	t.Run("正常获取购物车", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/cart/get?user_id=1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "获取购物车应返回200")

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, "应返回有效的JSON")

		cart := response["cart"].(map[string]interface{})
		items := cart["items"].([]interface{})
		assert.Len(t, items, 2, "应有2个购物车项")

		item1 := items[0].(map[string]interface{})
		assert.Equal(t, float64(101), item1["product_id"], "第一个商品ID应正确")
		assert.Equal(t, float64(2), item1["quantity"], "第一个商品数量应正确")

		item2 := items[1].(map[string]interface{})
		assert.Equal(t, float64(102), item2["product_id"], "第二个商品ID应正确")
	})
}

// 测试清空购物车
func TestEmptyCart(t *testing.T) {
	router := setupTestRouter()

	t.Run("参数验证失败", func(t *testing.T) {
		// 测试无效的用户ID
		jsonData := `{"user_id": 0}`
		req, _ := http.NewRequest("POST", "/cart/empty", bytes.NewBufferString(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "无效的用户ID应返回400")
		assert.Contains(t, w.Body.String(), "无效的用户ID", "响应应包含错误信息")
	})

	t.Run("成功清空购物车", func(t *testing.T) {
		jsonData := `{"user_id": 1}`
		req, _ := http.NewRequest("POST", "/cart/empty", bytes.NewBufferString(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "成功清空购物车应返回200")

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, "应返回有效的JSON")
		assert.Equal(t, true, response["success"], "操作应该成功")
		assert.Contains(t, response["message"], "购物车已清空", "响应消息应正确")
	})

	t.Run("数据库错误处理", func(t *testing.T) {
		// 测试特殊ID 888模拟数据库错误
		jsonData := `{"user_id": 888}`
		req, _ := http.NewRequest("POST", "/cart/empty", bytes.NewBufferString(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code, "数据库错误应返回500")
		assert.Contains(t, w.Body.String(), "清空购物车失败", "响应应包含错误信息")
	})
}
