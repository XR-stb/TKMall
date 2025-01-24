package service

import (
	"TKMall/build/proto_gen/auth"
	"TKMall/build/proto_gen/user"
	"context"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
)

func (s *UserServiceServer) Login(ctx context.Context, req *user.LoginReq) (*user.LoginResp, error) {
	userInfo, exists := s.Users[req.Email]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	err := bcrypt.CompareHashAndPassword([]byte(userInfo.Password), []byte(req.Password))
	if err != nil {
		return nil, fmt.Errorf("invalid password")
	}

	// 调用认证中心生成身份令牌
	tokenResp, err := s.AuthClient.DeliverTokenByRPC(ctx, &auth.DeliverTokenReq{UserId: userInfo.ID})
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %v", err)
	}

	log.Printf("Generated token for user %d: %s", userInfo.ID, tokenResp.Token)

	return &user.LoginResp{UserId: userInfo.ID}, nil
}
