package service

import (
	"TKMall/build/proto_gen/user"
	"TKMall/cmd/user/model"
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func (s *UserServiceServer) Register(ctx context.Context, req *user.RegisterReq) (*user.RegisterResp, error) {
	// if req.Password != req.ConfirmPassword {
	// 	return nil, fmt.Errorf("passwords do not match")
	// }

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}

	// 检查用户是否已经存在
	var existingUser model.User
	if err := s.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return nil, fmt.Errorf("user already exists")
	} else if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing user: %v", err)
	}

	var userID int64
	const maxRetries = 10
	for i := 0; i < maxRetries; i++ {
		// 使用雪花算法生成用户ID
		userID = s.Node.Generate().Int64()

		// 检查生成的ID是否已经存在
		var idCheckUser model.User
		if err := s.DB.Where("id = ?", userID).First(&idCheckUser).Error; err == gorm.ErrRecordNotFound {
			break
		}

		if i == maxRetries-1 {
			return nil, fmt.Errorf("failed to generate unique user ID after %d attempts", maxRetries)
		}
	}

	// 创建新用户
	newUser := &model.User{
		ID:       userID,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	if err := s.DB.Create(newUser).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	return &user.RegisterResp{UserId: int32(userID)}, nil
}
