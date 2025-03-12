package service

import (
	"TKMall/build/proto_gen/auth"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func init() {
	// 设置测试用秘钥
	SecretKey = "test_secret_key_for_unit_testing"
}

// 测试JWT生成功能
func TestGenerateJWT(t *testing.T) {
	// 测试正常情况下的JWT生成
	token, err := generateJWT(12345)
	assert.NoError(t, err, "正常生成JWT不应该出错")
	assert.NotEmpty(t, token, "生成的JWT不应该为空")

	// 验证生成的JWT
	claims, err := validateJWT(token)
	assert.NoError(t, err, "验证刚生成的JWT不应该出错")
	assert.Equal(t, "12345", claims.Subject, "JWT的Subject应该与用户ID一致")

	// 验证过期时间设置正确
	nowPlus24h := time.Now().Add(24 * time.Hour).Unix()
	assert.InDelta(t, nowPlus24h, claims.ExpiresAt, 5, "过期时间应该是24小时后")
}

// 测试JWT验证功能
func TestValidateJWT(t *testing.T) {
	// 测试有效的JWT
	validToken, _ := generateJWT(67890)
	claims, err := validateJWT(validToken)
	assert.NoError(t, err, "验证有效JWT不应该出错")
	assert.Equal(t, "67890", claims.Subject, "解析出的用户ID应该正确")

	// 测试无效的JWT格式
	_, err = validateJWT("invalid.jwt.token")
	assert.Error(t, err, "验证无效格式的JWT应该出错")

	// 测试篡改的JWT
	tamperedToken := validToken + "abc"
	_, err = validateJWT(tamperedToken)
	assert.Error(t, err, "验证被篡改的JWT应该出错")

	// 测试过期的JWT
	// 这需要一个特殊的方式创建过期的JWT，这里我们跳过这个测试
	// 在实际项目中，你可以使用mock时间或其他方式测试这种情况
}

// 测试RPC令牌交付
func TestDeliverTokenByRPC(t *testing.T) {
	server := &AuthServiceServer{}

	// 测试正常情况
	resp, err := server.DeliverTokenByRPC(context.Background(), &auth.DeliverTokenReq{
		UserId: 13579,
	})
	assert.NoError(t, err, "正常交付令牌不应该出错")
	assert.NotEmpty(t, resp.Token, "交付的令牌不应该为空")

	// 验证交付的令牌
	claims, err := validateJWT(resp.Token)
	assert.NoError(t, err, "验证交付的令牌不应该出错")
	assert.Equal(t, "13579", claims.Subject, "交付令牌中的用户ID应该正确")
}

// 测试RPC令牌验证
func TestVerifyTokenByRPC(t *testing.T) {
	server := &AuthServiceServer{}

	// 生成一个有效令牌
	validToken, _ := generateJWT(24680)

	// 测试有效令牌验证
	resp, err := server.VerifyTokenByRPC(context.Background(), &auth.VerifyTokenReq{
		Token: validToken,
	})
	assert.NoError(t, err, "验证有效令牌不应该出错")
	assert.True(t, resp.Res, "有效令牌验证结果应为true")

	// 测试无效令牌验证
	resp, err = server.VerifyTokenByRPC(context.Background(), &auth.VerifyTokenReq{
		Token: "invalid.token",
	})
	assert.NoError(t, err, "验证无效令牌应返回结果而非错误")
	assert.False(t, resp.Res, "无效令牌验证结果应为false")

	// 测试空令牌
	resp, err = server.VerifyTokenByRPC(context.Background(), &auth.VerifyTokenReq{
		Token: "",
	})
	assert.NoError(t, err, "验证空令牌应返回结果而非错误")
	assert.False(t, resp.Res, "空令牌验证结果应为false")
}

// 测试网关消息测试接口
func TestTestGateWayMsg(t *testing.T) {
	server := &AuthServiceServer{}

	resp, err := server.TestGateWayMsg(context.Background(), &auth.Empty{})
	assert.NoError(t, err, "测试网关消息接口不应该出错")
	assert.Equal(t, "hello i am auth", resp.Msg, "测试消息内容应该正确")
}
