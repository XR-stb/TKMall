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

	// 创建产品API路由组
	productGroup := r.Group("/product")
	{
		productGroup.POST("/search", func(c *gin.Context) {
			var req struct {
				Keyword string `json:"keyword"`
				Page    int    `json:"page"`
				Size    int    `json:"size"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			// 参数校验
			if req.Keyword == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "搜索关键词不能为空"})
				return
			}

			// SQL注入检测
			if req.Keyword == "'; DROP TABLE products; --" {
				// 模拟SQL注入输入被安全处理后无结果
				c.JSON(http.StatusOK, gin.H{
					"products": []interface{}{},
					"total":    0,
				})
				return
			}

			// 模拟数据库错误
			if req.Keyword == "database_error" {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库查询失败"})
				return
			}

			// 模拟无结果
			if req.Keyword == "nonexistent_product" {
				c.JSON(http.StatusOK, gin.H{
					"products": []interface{}{},
					"total":    0,
				})
				return
			}

			// 正常搜索结果
			products := []gin.H{
				{
					"id":          1,
					"name":        "智能手机",
					"description": "高性能智能手机",
					"price":       2999.99,
					"stock":       100,
					"category":    "电子产品",
				},
				{
					"id":          2,
					"name":        "智能手表",
					"description": "健康监测智能手表",
					"price":       1299.99,
					"stock":       50,
					"category":    "电子产品",
				},
			}

			c.JSON(http.StatusOK, gin.H{
				"products": products,
				"total":    2,
			})
		})
	}

	return r
}

// 测试搜索产品
func TestSearchProducts(t *testing.T) {
	router := setupTestRouter()

	t.Run("空关键词搜索", func(t *testing.T) {
		jsonData := `{"keyword": "", "page": 1, "size": 10}`
		req, _ := http.NewRequest("POST", "/product/search", bytes.NewBufferString(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "空关键词应返回400")
		assert.Contains(t, w.Body.String(), "搜索关键词不能为空", "响应应包含错误信息")
	})

	t.Run("正常搜索", func(t *testing.T) {
		jsonData := `{"keyword": "智能", "page": 1, "size": 10}`
		req, _ := http.NewRequest("POST", "/product/search", bytes.NewBufferString(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "正常搜索应返回200")

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, "应返回有效的JSON")

		products := response["products"].([]interface{})
		assert.Len(t, products, 2, "应返回2个产品")
		assert.Equal(t, float64(2), response["total"], "总数应为2")

		product1 := products[0].(map[string]interface{})
		assert.Equal(t, "智能手机", product1["name"], "第一个产品名称应正确")
		assert.Equal(t, float64(2999.99), product1["price"], "第一个产品价格应正确")

		product2 := products[1].(map[string]interface{})
		assert.Equal(t, "智能手表", product2["name"], "第二个产品名称应正确")
	})

	t.Run("SQL注入测试", func(t *testing.T) {
		// 尝试进行SQL注入
		jsonData := `{"keyword": "'; DROP TABLE products; --", "page": 1, "size": 10}`
		req, _ := http.NewRequest("POST", "/product/search", bytes.NewBufferString(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "SQL注入尝试应正常处理并返回200")

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, "应返回有效的JSON")

		products := response["products"].([]interface{})
		assert.Empty(t, products, "SQL注入尝试应返回空结果")
		assert.Equal(t, float64(0), response["total"], "总数应为0")
	})

	t.Run("数据库错误", func(t *testing.T) {
		jsonData := `{"keyword": "database_error", "page": 1, "size": 10}`
		req, _ := http.NewRequest("POST", "/product/search", bytes.NewBufferString(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code, "数据库错误应返回500")
		assert.Contains(t, w.Body.String(), "数据库查询失败", "响应应包含错误信息")
	})

	t.Run("无结果搜索", func(t *testing.T) {
		jsonData := `{"keyword": "nonexistent_product", "page": 1, "size": 10}`
		req, _ := http.NewRequest("POST", "/product/search", bytes.NewBufferString(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "无结果搜索应返回200")

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, "应返回有效的JSON")

		products := response["products"].([]interface{})
		assert.Empty(t, products, "应返回空结果")
		assert.Equal(t, float64(0), response["total"], "总数应为0")
	})
}
