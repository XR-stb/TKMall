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

	// 创建支付API路由组
	paymentGroup := r.Group("/payment")
	{
		paymentGroup.POST("/charge", func(c *gin.Context) {
			var req struct {
				OrderID        string  `json:"order_id"`
				UserID         int64   `json:"user_id"`
				Amount         float64 `json:"amount"`
				CardNumber     string  `json:"card_number"`
				CardExpiry     string  `json:"card_expiry"`
				CardCVV        string  `json:"card_cvv"`
				CardHolderName string  `json:"card_holder_name"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			// 参数校验
			if req.OrderID == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "订单ID不能为空"})
				return
			}

			if req.UserID <= 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID无效"})
				return
			}

			if req.Amount <= 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "支付金额必须大于0"})
				return
			}

			if req.CardNumber == "" || req.CardExpiry == "" || req.CardCVV == "" || req.CardHolderName == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "信用卡信息不完整"})
				return
			}

			// 信用卡号格式验证
			if len(req.CardNumber) < 13 || len(req.CardNumber) > 19 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "信用卡号格式不正确"})
				return
			}

			// 模拟支付失败
			if req.OrderID == "fail_payment" {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "支付处理失败: 资金不足"})
				return
			}

			// 模拟特定卡号处理
			cardType := "未知"
			switch {
			case req.CardNumber[:1] == "4":
				cardType = "Visa"
			case req.CardNumber[:1] == "5":
				cardType = "MasterCard"
			case req.CardNumber[:2] == "34" || req.CardNumber[:2] == "37":
				cardType = "American Express"
			}

			// 成功处理支付
			c.JSON(http.StatusOK, gin.H{
				"transaction_id": "txn_" + req.OrderID,
				"status":         "成功",
				"card_type":      cardType,
				"card_last_four": req.CardNumber[len(req.CardNumber)-4:],
				"amount":         req.Amount,
			})
		})
	}

	return r
}

// 测试支付功能
func TestCharge(t *testing.T) {
	router := setupTestRouter()

	t.Run("参数验证", func(t *testing.T) {
		// 测试空订单ID
		jsonData := `{
			"order_id": "",
			"user_id": 1,
			"amount": 100.0,
			"card_number": "4111111111111111",
			"card_expiry": "12/25",
			"card_cvv": "123",
			"card_holder_name": "测试用户"
		}`
		req, _ := http.NewRequest("POST", "/payment/charge", bytes.NewBufferString(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "空订单ID应返回400")
		assert.Contains(t, w.Body.String(), "订单ID不能为空", "响应应包含错误信息")

		// 测试无效用户ID
		jsonData = `{
			"order_id": "order123",
			"user_id": 0,
			"amount": 100.0,
			"card_number": "4111111111111111",
			"card_expiry": "12/25",
			"card_cvv": "123",
			"card_holder_name": "测试用户"
		}`
		req, _ = http.NewRequest("POST", "/payment/charge", bytes.NewBufferString(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "无效用户ID应返回400")
		assert.Contains(t, w.Body.String(), "用户ID无效", "响应应包含错误信息")

		// 测试无效金额
		jsonData = `{
			"order_id": "order123",
			"user_id": 1,
			"amount": 0,
			"card_number": "4111111111111111",
			"card_expiry": "12/25",
			"card_cvv": "123",
			"card_holder_name": "测试用户"
		}`
		req, _ = http.NewRequest("POST", "/payment/charge", bytes.NewBufferString(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "无效金额应返回400")
		assert.Contains(t, w.Body.String(), "支付金额必须大于0", "响应应包含错误信息")

		// 测试缺失卡信息
		jsonData = `{
			"order_id": "order123",
			"user_id": 1,
			"amount": 100.0,
			"card_number": "",
			"card_expiry": "12/25",
			"card_cvv": "123",
			"card_holder_name": "测试用户"
		}`
		req, _ = http.NewRequest("POST", "/payment/charge", bytes.NewBufferString(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "缺失卡信息应返回400")
		assert.Contains(t, w.Body.String(), "信用卡信息不完整", "响应应包含错误信息")
	})

	t.Run("信用卡验证", func(t *testing.T) {
		// 测试无效卡号（太短）
		jsonData := `{
			"order_id": "order123",
			"user_id": 1,
			"amount": 100.0,
			"card_number": "4111",
			"card_expiry": "12/25",
			"card_cvv": "123",
			"card_holder_name": "测试用户"
		}`
		req, _ := http.NewRequest("POST", "/payment/charge", bytes.NewBufferString(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "无效卡号应返回400")
		assert.Contains(t, w.Body.String(), "信用卡号格式不正确", "响应应包含错误信息")
	})

	t.Run("成功支付", func(t *testing.T) {
		// 测试有效支付
		jsonData := `{
			"order_id": "order123",
			"user_id": 1,
			"amount": 100.0,
			"card_number": "4111111111111111",
			"card_expiry": "12/25",
			"card_cvv": "123",
			"card_holder_name": "测试用户"
		}`
		req, _ := http.NewRequest("POST", "/payment/charge", bytes.NewBufferString(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "成功支付应返回200")

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, "应返回有效的JSON")

		assert.Equal(t, "txn_order123", response["transaction_id"], "交易ID应正确")
		assert.Equal(t, "成功", response["status"], "支付状态应为成功")
		assert.Equal(t, "Visa", response["card_type"], "卡类型应正确")
		assert.Equal(t, "1111", response["card_last_four"], "卡末四位应正确")
		assert.Equal(t, float64(100.0), response["amount"], "支付金额应正确")
	})

	t.Run("支付失败", func(t *testing.T) {
		// 测试支付失败
		jsonData := `{
			"order_id": "fail_payment",
			"user_id": 1,
			"amount": 100.0,
			"card_number": "4111111111111111",
			"card_expiry": "12/25",
			"card_cvv": "123",
			"card_holder_name": "测试用户"
		}`
		req, _ := http.NewRequest("POST", "/payment/charge", bytes.NewBufferString(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code, "支付失败应返回500")
		assert.Contains(t, w.Body.String(), "资金不足", "响应应包含具体失败原因")
	})

	t.Run("不同卡类型处理", func(t *testing.T) {
		// 测试MasterCard
		jsonData := `{
			"order_id": "order456",
			"user_id": 1,
			"amount": 200.0,
			"card_number": "5555555555554444",
			"card_expiry": "12/25",
			"card_cvv": "123",
			"card_holder_name": "测试用户"
		}`
		req, _ := http.NewRequest("POST", "/payment/charge", bytes.NewBufferString(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "MasterCard支付应返回200")

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "MasterCard", response["card_type"], "卡类型应为MasterCard")
		assert.Equal(t, "4444", response["card_last_four"], "卡末四位应正确")

		// 测试American Express
		jsonData = `{
			"order_id": "order789",
			"user_id": 1,
			"amount": 300.0,
			"card_number": "371449635398431",
			"card_expiry": "12/25",
			"card_cvv": "1234",
			"card_holder_name": "测试用户"
		}`
		req, _ = http.NewRequest("POST", "/payment/charge", bytes.NewBufferString(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "American Express支付应返回200")

		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "American Express", response["card_type"], "卡类型应为American Express")
		assert.Equal(t, "8431", response["card_last_four"], "卡末四位应正确")
	})
}

// 测试获取信用卡类型函数
func TestDetectCardType(t *testing.T) {
	tests := []struct {
		cardNumber string
		expected   string
	}{
		{"4111111111111111", "Visa"},
		{"5555555555554444", "MasterCard"},
		{"371449635398431", "American Express"},
		{"6011111111111117", "未知"},
	}

	for _, test := range tests {
		router := setupTestRouter()

		jsonData := `{
			"order_id": "order123",
			"user_id": 1,
			"amount": 100.0,
			"card_number": "` + test.cardNumber + `",
			"card_expiry": "12/25",
			"card_cvv": "123",
			"card_holder_name": "测试用户"
		}`

		req, _ := http.NewRequest("POST", "/payment/charge", bytes.NewBufferString(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, test.expected, response["card_type"], "卡类型检测应正确: "+test.cardNumber)
	}
}

// 测试获取信用卡末尾4位函数
func TestGetLastFourDigits(t *testing.T) {
	tests := []struct {
		cardNumber string
		expected   string
	}{
		{"4111111111111111", "1111"},
		{"5555555555554444", "4444"},
		{"371449635398431", "8431"},
		{"6011111111111117", "1117"},
	}

	for _, test := range tests {
		router := setupTestRouter()

		jsonData := `{
			"order_id": "order123",
			"user_id": 1,
			"amount": 100.0,
			"card_number": "` + test.cardNumber + `",
			"card_expiry": "12/25",
			"card_cvv": "123",
			"card_holder_name": "测试用户"
		}`

		req, _ := http.NewRequest("POST", "/payment/charge", bytes.NewBufferString(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, test.expected, response["card_last_four"], "卡末四位提取应正确: "+test.cardNumber)
	}
}
