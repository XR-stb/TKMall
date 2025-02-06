package service

import (
	"TKMall/build/proto_gen/auth"
	"TKMall/build/proto_gen/user"
	"TKMall/cmd/user/model"
	"context"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Login service
func (s *UserServiceServer) Login(ctx context.Context, req *user.LoginReq) (*user.LoginResp, error) {
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

	// generate token
	tokenResp, err := s.AuthClient.DeliverTokenByRPC(ctx, &auth.DeliverTokenReq{UserId: int32(userInfo.ID)})
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %v", err)
	}

	log.Printf("Generated token for user %d: %s", userInfo.ID, tokenResp.Token)

	// may need to return token
	// return &user.LoginResp{UserId: int32(userInfo.ID), Token: tokenResp.Token}, nil
	return &user.LoginResp{UserId: int32(userInfo.ID)}, nil
}
