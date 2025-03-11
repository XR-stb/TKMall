package service

import (
	"context"

	"TKMall/build/proto_gen/cart"
	"TKMall/cmd/cart/model"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// EmptyCart 清空用户的购物车
func (s *CartServiceServer) EmptyCart(ctx context.Context, req *cart.EmptyCartReq) (*cart.EmptyCartResp, error) {
	// 参数校验
	if req.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "用户ID不能为空")
	}

	// 查询用户的购物车
	var userCart model.Cart
	if err := s.DB.Where("user_id = ?", req.UserId).First(&userCart).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 用户没有购物车，直接返回成功
			return &cart.EmptyCartResp{}, nil
		}
		return nil, status.Errorf(codes.Internal, "查询购物车失败: %v", err)
	}

	// 开启事务，删除所有购物车项
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		// 删除该购物车下的所有商品
		if err := tx.Where("cart_id = ?", userCart.ID).Delete(&model.CartItem{}).Error; err != nil {
			return err
		}

		// 更新购物车最后修改时间
		if err := tx.Model(&userCart).Update("updated_at", gorm.Expr("NOW()")).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, status.Errorf(codes.Internal, "清空购物车失败: %v", err)
	}

	return &cart.EmptyCartResp{}, nil
}
