package service

import (
	"context"

	"TKMall/build/proto_gen/cart"
	"TKMall/cmd/cart/model"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// GetCart 获取用户的购物车
func (s *CartServiceServer) GetCart(ctx context.Context, req *cart.GetCartReq) (*cart.GetCartResp, error) {
	// 参数校验
	if req.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "用户ID不能为空")
	}

	// 查询用户的购物车
	var userCart model.Cart
	if err := s.DB.Where("user_id = ?", req.UserId).First(&userCart).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 用户没有购物车，返回空购物车
			return &cart.GetCartResp{
				Cart: &cart.Cart{
					UserId: req.UserId,
					Items:  []*cart.CartItem{},
				},
			}, nil
		}
		return nil, status.Errorf(codes.Internal, "获取购物车失败: %v", err)
	}

	// 查询购物车中的商品
	var cartItems []model.CartItem
	if err := s.DB.Where("cart_id = ?", userCart.ID).Find(&cartItems).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "获取购物车商品失败: %v", err)
	}

	// 转换成Proto对象
	protoItems := make([]*cart.CartItem, 0, len(cartItems))
	for _, item := range cartItems {
		protoItems = append(protoItems, &cart.CartItem{
			ProductId: uint32(item.ProductID),
			Quantity:  int32(item.Quantity),
		})
	}

	return &cart.GetCartResp{
		Cart: &cart.Cart{
			UserId: req.UserId,
			Items:  protoItems,
		},
	}, nil
}
