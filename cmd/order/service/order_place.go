package service

import (
	"context"
	"fmt"

	"TKMall/build/proto_gen/cart"
	"TKMall/build/proto_gen/order"
	"TKMall/cmd/order/model"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// PlaceOrder 创建订单
func (s *OrderServiceServer) PlaceOrder(ctx context.Context, req *order.PlaceOrderReq) (*order.PlaceOrderResp, error) {
	// 参数校验
	if req.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "用户ID不能为空")
	}
	if req.Address == nil {
		return nil, status.Error(codes.InvalidArgument, "收货地址不能为空")
	}
	if len(req.OrderItems) == 0 {
		return nil, status.Error(codes.InvalidArgument, "订单项不能为空")
	}

	// 生成唯一的订单ID
	orderID := fmt.Sprintf("ORD-%s", s.Node.Generate().String())

	// 计算订单总金额
	var totalAmount float64
	for _, item := range req.OrderItems {
		totalAmount += float64(item.Cost)
	}

	// 创建订单记录
	orderInfo := model.Order{
		OrderID:      orderID,
		UserID:       req.UserId,
		Status:       model.OrderStatusCreated,
		TotalAmount:  totalAmount,
		Address:      convertAddressToModel(req.Address),
		Email:        req.Email,
		UserCurrency: req.UserCurrency,
	}

	// 创建订单项
	var orderItems []model.OrderItem
	for _, item := range req.OrderItems {
		if item.Item == nil {
			continue
		}
		orderItem := model.OrderItem{
			OrderID:    orderID,
			ProductID:  uint(item.Item.ProductId),
			Quantity:   int(item.Item.Quantity),
			Price:      float64(item.Cost) / float64(item.Item.Quantity), // 单价 = 总价 / 数量
			TotalPrice: float64(item.Cost),
		}
		orderItems = append(orderItems, orderItem)
	}

	// 使用事务保证数据一致性
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		// 保存订单
		if err := tx.Create(&orderInfo).Error; err != nil {
			return fmt.Errorf("创建订单失败: %w", err)
		}

		// 保存订单项
		if err := tx.CreateInBatches(orderItems, 100).Error; err != nil {
			return fmt.Errorf("创建订单项失败: %w", err)
		}

		// 清空购物车（可选，取决于业务需求）
		emptyCartReq := &cart.EmptyCartReq{
			UserId: req.UserId,
		}

		_, err := s.Proxy.Call(ctx, "cart", "EmptyCart", emptyCartReq)
		if err != nil {
			// 仅记录日志，不影响订单创建
			return fmt.Errorf("清空购物车失败: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, status.Errorf(codes.Internal, "创建订单失败: %v", err)
	}

	// 发布订单创建事件
	if s.EventBus != nil {
		// 这里可以发布订单创建事件，供其他服务订阅处理
		// s.EventBus.Publish("order.created", orderEvent)
	}

	return &order.PlaceOrderResp{
		Order: &order.OrderResult{
			OrderId: orderID,
		},
	}, nil
}

// 将proto的Address转换为model的Address
func convertAddressToModel(addr *order.Address) model.Address {
	if addr == nil {
		return model.Address{}
	}
	return model.Address{
		StreetAddress: addr.StreetAddress,
		City:          addr.City,
		State:         addr.State,
		Country:       addr.Country,
		ZipCode:       addr.ZipCode,
	}
}
