package service

import (
	"context"
	"fmt"

	"TKMall/build/proto_gen/cart"
	"TKMall/cmd/cart/model"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// AddItem 添加商品到购物车
func (s *CartServiceServer) AddItem(ctx context.Context, req *cart.AddItemReq) (*cart.AddItemResp, error) {
	// 参数校验
	if req.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "用户ID不能为空")
	}
	if req.Item == nil || req.Item.ProductId == 0 {
		return nil, status.Error(codes.InvalidArgument, "商品信息不完整")
	}
	if req.Item.Quantity <= 0 {
		return nil, status.Error(codes.InvalidArgument, "商品数量必须大于0")
	}

	// 查询或创建用户的购物车
	var userCart model.Cart
	if err := s.DB.Where("user_id = ?", req.UserId).FirstOrCreate(&userCart, model.Cart{
		UserID: uint(req.UserId),
	}).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "获取购物车失败: %v", err)
	}

	// 查询该商品价格（实际项目中需要调用产品服务获取最新价格）
	// 这里简化处理，假设我们从Product服务获取了价格
	productPrice := 0.0 // 默认价格，实际项目应该从product服务获取

	// TODO: 调用产品服务获取商品信息和价格
	// productInfo, err := s.Proxy.GetProduct(ctx, req.Item.ProductId)
	// if err != nil {
	//     return nil, status.Errorf(codes.NotFound, "商品不存在或已下架: %v", err)
	// }
	// productPrice = productInfo.Price

	// 查找购物车中是否已有该商品
	var cartItem model.CartItem
	result := s.DB.Where("cart_id = ? AND product_id = ?", userCart.ID, req.Item.ProductId).First(&cartItem)

	err := s.DB.Transaction(func(tx *gorm.DB) error {
		if result.Error == nil {
			// 商品已存在，更新数量
			cartItem.Quantity = int(req.Item.Quantity)
			if err := tx.Save(&cartItem).Error; err != nil {
				return fmt.Errorf("更新购物车项失败: %v", err)
			}
		} else if result.Error == gorm.ErrRecordNotFound {
			// 商品不存在，创建新的购物车项
			newItem := model.CartItem{
				CartID:    userCart.ID,
				ProductID: uint(req.Item.ProductId),
				Quantity:  int(req.Item.Quantity),
				Price:     productPrice,
			}
			if err := tx.Create(&newItem).Error; err != nil {
				return fmt.Errorf("添加购物车项失败: %v", err)
			}
		} else {
			return fmt.Errorf("查询购物车项失败: %v", result.Error)
		}

		// 更新购物车最后修改时间
		if err := tx.Model(&userCart).Update("updated_at", gorm.Expr("NOW()")).Error; err != nil {
			return fmt.Errorf("更新购物车时间失败: %v", err)
		}

		return nil
	})

	if err != nil {
		return nil, status.Errorf(codes.Internal, "操作购物车失败: %v", err)
	}

	return &cart.AddItemResp{}, nil
}
