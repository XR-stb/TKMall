package service

import (
	"context"

	"TKMall/build/proto_gen/cart"
	"TKMall/build/proto_gen/checkout"
	"TKMall/build/proto_gen/order"
	"TKMall/build/proto_gen/payment"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Checkout 处理结账请求
func (s *CheckoutServiceServer) Checkout(ctx context.Context, req *checkout.CheckoutReq) (*checkout.CheckoutResp, error) {
	// 参数校验
	if req.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "用户ID不能为空")
	}
	if req.Address == nil {
		return nil, status.Error(codes.InvalidArgument, "收货地址不能为空")
	}
	if req.CreditCard == nil {
		return nil, status.Error(codes.InvalidArgument, "支付信息不能为空")
	}

	// 1. 获取用户购物车
	cartReq := &cart.GetCartReq{UserId: req.UserId}
	cartRespInterface, err := s.Proxy.Call(ctx, "cart", "GetCart", cartReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取购物车失败: %v", err)
	}

	cartResp, ok := cartRespInterface.(*cart.GetCartResp)
	if !ok {
		return nil, status.Error(codes.Internal, "响应类型转换失败")
	}

	if cartResp.Cart == nil || len(cartResp.Cart.Items) == 0 {
		return nil, status.Error(codes.FailedPrecondition, "购物车为空，无法结账")
	}

	// 2. 创建订单项
	var orderItems []*order.OrderItem
	for _, item := range cartResp.Cart.Items {
		// 实际应用中，应该从商品服务获取商品价格
		itemCost := float32(item.Quantity) * 100.0 // 假设每个商品100元
		orderItem := &order.OrderItem{
			Item: &cart.CartItem{
				ProductId: item.ProductId,
				Quantity:  item.Quantity,
			},
			Cost: itemCost,
		}
		orderItems = append(orderItems, orderItem)
	}

	// 3. 创建订单
	orderReq := &order.PlaceOrderReq{
		UserId:       req.UserId,
		UserCurrency: "CNY", // 默认使用人民币
		Address:      convertProtoAddress(req.Address),
		Email:        req.Email,
		OrderItems:   orderItems,
	}

	orderRespInterface, err := s.Proxy.Call(ctx, "order", "PlaceOrder", orderReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "创建订单失败: %v", err)
	}

	orderResp, ok := orderRespInterface.(*order.PlaceOrderResp)
	if !ok {
		return nil, status.Error(codes.Internal, "响应类型转换失败")
	}

	if orderResp.Order == nil || orderResp.Order.OrderId == "" {
		return nil, status.Error(codes.Internal, "创建订单失败：无效的订单ID")
	}

	// 4. 处理支付
	// 计算订单总金额
	totalAmount := float32(0)
	for _, item := range orderItems {
		totalAmount += item.Cost
	}

	paymentReq := &payment.ChargeReq{
		Amount: totalAmount,
		CreditCard: &payment.CreditCardInfo{
			CreditCardNumber:          req.CreditCard.CreditCardNumber,
			CreditCardCvv:             req.CreditCard.CreditCardCvv,
			CreditCardExpirationYear:  req.CreditCard.CreditCardExpirationYear,
			CreditCardExpirationMonth: req.CreditCard.CreditCardExpirationMonth,
		},
		OrderId: orderResp.Order.OrderId,
		UserId:  req.UserId,
	}

	paymentRespInterface, err := s.Proxy.Call(ctx, "payment", "Charge", paymentReq)
	if err != nil {
		// 支付失败，但订单已创建
		return nil, status.Errorf(codes.Internal, "支付处理失败: %v", err)
	}

	paymentResp, ok := paymentRespInterface.(*payment.ChargeResp)
	if !ok {
		return nil, status.Error(codes.Internal, "响应类型转换失败")
	}

	// 5. 返回结果
	return &checkout.CheckoutResp{
		OrderId:       orderResp.Order.OrderId,
		TransactionId: paymentResp.TransactionId,
	}, nil
}
