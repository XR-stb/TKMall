package service

import (
	"context"

	"TKMall/build/proto_gen/cart"
	"TKMall/build/proto_gen/order"
	"TKMall/cmd/order/model"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ListOrder 查询用户订单列表
func (s *OrderServiceServer) ListOrder(ctx context.Context, req *order.ListOrderReq) (*order.ListOrderResp, error) {
	// 参数校验
	if req.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "用户ID不能为空")
	}

	// 查询用户订单
	var orders []model.Order
	if err := s.DB.Where("user_id = ?", req.UserId).Order("created_at DESC").Find(&orders).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "查询订单失败: %v", err)
	}

	// 如果没有找到订单，返回空列表
	if len(orders) == 0 {
		return &order.ListOrderResp{Orders: []*order.Order{}}, nil
	}

	// 转换为proto格式
	protoOrders := make([]*order.Order, 0, len(orders))
	for _, ord := range orders {
		// 查询订单项
		var orderItems []model.OrderItem
		if err := s.DB.Where("order_id = ?", ord.OrderID).Find(&orderItems).Error; err != nil {
			return nil, status.Errorf(codes.Internal, "查询订单项失败: %v", err)
		}

		// 转换订单项
		protoOrderItems := make([]*order.OrderItem, 0, len(orderItems))
		for _, item := range orderItems {
			protoOrderItems = append(protoOrderItems, &order.OrderItem{
				Item: &cart.CartItem{
					ProductId: uint32(item.ProductID),
					Quantity:  int32(item.Quantity),
				},
				Cost: float32(item.TotalPrice),
			})
		}

		// 创建订单对象
		protoOrder := &order.Order{
			OrderId:      ord.OrderID,
			UserId:       ord.UserID,
			UserCurrency: ord.UserCurrency,
			Email:        ord.Email,
			OrderItems:   protoOrderItems,
			Address:      convertAddressToProto(ord.Address),
			CreatedAt:    int32(ord.CreatedAt.Unix()),
		}

		protoOrders = append(protoOrders, protoOrder)
	}

	return &order.ListOrderResp{
		Orders: protoOrders,
	}, nil
}

// 将model的Address转换为proto的Address
func convertAddressToProto(addr model.Address) *order.Address {
	return &order.Address{
		StreetAddress: addr.StreetAddress,
		City:          addr.City,
		State:         addr.State,
		Country:       addr.Country,
		ZipCode:       addr.ZipCode,
	}
}
