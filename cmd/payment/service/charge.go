package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"TKMall/build/proto_gen/order"
	"TKMall/build/proto_gen/payment"
	"TKMall/cmd/payment/model"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Charge 处理支付请求
func (s *PaymentServiceServer) Charge(ctx context.Context, req *payment.ChargeReq) (*payment.ChargeResp, error) {
	// 参数校验
	if req.OrderId == "" {
		return nil, status.Error(codes.InvalidArgument, "订单ID不能为空")
	}
	if req.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "用户ID不能为空")
	}
	if req.Amount <= 0 {
		return nil, status.Error(codes.InvalidArgument, "支付金额必须大于0")
	}
	if req.CreditCard == nil {
		return nil, status.Error(codes.InvalidArgument, "支付信息不能为空")
	}

	// 生成唯一的交易ID
	transactionID := fmt.Sprintf("TXN-%s", s.Node.Generate().String())

	// 获取信用卡信息
	cardNumber := req.CreditCard.CreditCardNumber
	// CVV在真实支付场景中会被使用，这里仅做演示
	// cardCvv := req.CreditCard.CreditCardCvv
	cardExpMonth := req.CreditCard.CreditCardExpirationMonth
	cardExpYear := req.CreditCard.CreditCardExpirationYear

	// 卡号基本验证
	if len(cardNumber) < 13 || len(cardNumber) > 19 {
		return nil, status.Error(codes.InvalidArgument, "无效的信用卡号")
	}

	// 获取卡的后四位和卡类型
	lastFourDigits := getLastFourDigits(cardNumber)
	cardType := detectCardType(cardNumber)

	// 创建账单地址哈希（实际环境中应包含完整地址信息）
	billingAddressHash := fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%d", req.UserId))))

	// 保存信用卡信息（实际环境中应该使用专业的PCI合规存储）
	creditCard := model.CreditCard{
		UserID:             req.UserId,
		LastFourDigits:     lastFourDigits,
		ExpirationMonth:    int(cardExpMonth),
		ExpirationYear:     int(cardExpYear),
		CardType:           cardType,
		BillingAddressHash: billingAddressHash,
	}

	// 在真实环境中，这里应该调用实际的支付网关API
	// mockGatewayResponse := callPaymentGateway(cardNumber, cardCvv, cardExpMonth, cardExpYear, req.Amount)

	// 模拟支付处理（实际环境中应该使用真实的支付网关）
	// 为了演示，我们假设支付总是成功的
	paymentSuccessful := true
	errorCode := ""
	errorMessage := ""

	// 记录交易信息
	now := time.Now()
	transaction := model.Transaction{
		TransactionID:   transactionID,
		OrderID:         req.OrderId,
		UserID:          req.UserId,
		Amount:          float64(req.Amount),
		Status:          model.PaymentStatusPending,
		PaymentMethod:   PaymentMethodCreditCard,
		LastFourDigits:  lastFourDigits,
		TransactionTime: &now,
	}

	// 模拟支付网关响应
	gatewayResponse := fmt.Sprintf(`{"success":%t,"transaction_id":"%s","time":"%s"}`,
		paymentSuccessful, transactionID, now.Format(time.RFC3339))
	transaction.GatewayResponseRaw = gatewayResponse

	// 根据支付结果更新交易状态
	if paymentSuccessful {
		transaction.Status = model.PaymentStatusCompleted
	} else {
		transaction.Status = model.PaymentStatusFailed
		transaction.ErrorCode = errorCode
		transaction.ErrorMessage = errorMessage
		return nil, status.Errorf(codes.FailedPrecondition, "支付失败: %s", errorMessage)
	}

	// 保存信用卡和交易信息
	if err := s.DB.Create(&creditCard).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "保存信用卡信息失败: %v", err)
	}
	transaction.CreditCardID = creditCard.ID

	if err := s.DB.Create(&transaction).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "保存交易记录失败: %v", err)
	}

	// 调用订单服务标记订单为已支付
	if transaction.Status == model.PaymentStatusCompleted {
		_, err := s.OrderService.MarkOrderPaid(ctx, &order.MarkOrderPaidReq{
			UserId:  req.UserId,
			OrderId: req.OrderId,
		})
		if err != nil {
			// 仅记录日志，不影响支付结果
			return nil, status.Errorf(codes.Internal, "标记订单已支付失败: %v", err)
		}
	}

	// 发布支付成功事件
	if s.EventBus != nil && transaction.Status == model.PaymentStatusCompleted {
		// 在实际项目中可以发布支付成功事件
		// s.EventBus.Publish("payment.completed", paymentEvent)
	}

	return &payment.ChargeResp{
		TransactionId: transactionID,
	}, nil
}
