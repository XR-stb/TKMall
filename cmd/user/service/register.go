package service

import (
	"TKMall/build/proto_gen/user"
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func (s *UserServiceServer) Register(ctx context.Context, req *user.RegisterReq) (*user.RegisterResp, error) {
	// if req.Password != req.ConfirmPassword {
	// 	return nil, fmt.Errorf("passwords do not match")
	// }

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}

	if s.Users == nil {
		s.Users = make(map[string]*User)
	}

	if _, ok := s.Users[req.Email]; ok {
		return nil, fmt.Errorf("user already exist")
	}

	s.Users[req.Email] = &User{
		ID:       userIDCounter,
		Email:    req.Email,
		Password: string(hashedPassword),
	}
	userIDCounter++

	return &user.RegisterResp{UserId: s.Users[req.Email].ID}, nil
}

