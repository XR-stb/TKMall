package service

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

func handleListError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return status.Error(codes.NotFound, "未找到商品")
	}
	return status.Error(codes.Internal, "内部服务错误")
}

func handleGetError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return status.Error(codes.NotFound, "商品不存在")
	}
	return status.Error(codes.Internal, "内部服务错误")
}

func handleSearchError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return status.Error(codes.NotFound, "未找到匹配商品")
	}
	return status.Error(codes.Internal, "搜索服务暂时不可用")
}
