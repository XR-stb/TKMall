package service

import (
	"TKMall/build/proto_gen/auth"
	"TKMall/build/proto_gen/user"
	"TKMall/cmd/user/model"
	"TKMall/common/log"
	"context"
	"fmt"
	"runtime/debug"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

func loginParamCheck(input string) bool {
	if input == "" {
		return false
	}

	if len(input) < 6 {
		return false
	}

	// TODO: more check
	return true
}

// Login service
func (s *UserServiceServer) Login(ctx context.Context, req *user.LoginReq) (resp *user.LoginResp, err error) {
	// 添加panic恢复机制
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Login处理panic: %v\n堆栈: %s", r, debug.Stack())
			err = status.Error(codes.Internal, "服务内部错误")
			resp = nil
		}
	}()

	if !loginParamCheck(req.Email) || !loginParamCheck(req.Password) {
		log.Debugf("login req email: %s, password: %s invalid", req.Email, req.Password)
		return nil, fmt.Errorf("login param error")
	}
	var userInfo model.User

	if err := s.DB.Where("email = ?", req.Email).First(&userInfo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to retrieve user: %v", err)
	}

	// verify password
	err = bcrypt.CompareHashAndPassword([]byte(userInfo.Password), []byte(req.Password))
	if err != nil {
		return nil, fmt.Errorf("invalid password")
	}

	// 通过代理层调用 Auth 服务
	tokenReq := &auth.DeliverTokenReq{UserId: userInfo.ID}
	log.Infof("调用认证服务生成Token，请求数据: %+v", tokenReq)

	tokenResp, err := s.Proxy.Call(ctx, "auth", "DeliverTokenByRPC", tokenReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %v", err)
	}

	log.Infof("认证服务返回，响应类型: %T, 响应数据: %+v", tokenResp, tokenResp)

	// 使用类型断言，并检查断言是否成功
	authResp, ok := tokenResp.(*auth.DeliveryResp)
	if !ok {
		// 尝试从map[string]interface{}中提取token
		if mapResp, isMap := tokenResp.(map[string]interface{}); isMap {
			if token, exists := mapResp["token"]; exists {
				if tokenStr, isString := token.(string); isString {
					log.Infof("从map中成功提取token: %s", tokenStr)
					return &user.LoginResp{UserId: userInfo.ID, Token: tokenStr}, nil
				}
			}
		}

		log.Errorf("类型不匹配: 预期类型 *auth.DeliveryResp, 实际类型 %T, 值: %+v", tokenResp, tokenResp)
		return nil, status.Error(codes.Internal, "认证服务返回格式错误")
	}

	log.Infof("Generated token for user %d: %s", userInfo.ID, authResp.Token)

	return &user.LoginResp{UserId: userInfo.ID, Token: authResp.Token}, nil
}
