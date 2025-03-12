package service

import (
	"TKMall/build/proto_gen/auth"
	"TKMall/build/proto_gen/user"
	"TKMall/cmd/user/model"
	"TKMall/common/log"
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"
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
func (s *UserServiceServer) Login(ctx context.Context, req *user.LoginReq) (*user.LoginResp, error) {
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
	err := bcrypt.CompareHashAndPassword([]byte(userInfo.Password), []byte(req.Password))
	if err != nil {
		return nil, fmt.Errorf("invalid password")
	}

	// 通过代理层调用 Auth 服务
	tokenReq := &auth.DeliverTokenReq{UserId: int64(userInfo.ID)}
	tokenResp, err := s.Proxy.Call(ctx, "auth", "DeliverTokenByRPC", tokenReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %v", err)
	}

	authResp := tokenResp.(*auth.DeliveryResp)
	log.Infof("Generated token for user %d: %s", userInfo.ID, authResp.Token)

	return &user.LoginResp{UserId: userInfo.ID, Token: authResp.Token}, nil
}
