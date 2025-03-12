package service

import (
	"TKMall/build/proto_gen/order"
	"TKMall/build/proto_gen/payment"
	"TKMall/common/events"
	"TKMall/common/proxy"

	"github.com/bwmarrin/snowflake"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// PaymentServiceServer 支付服务实现
type PaymentServiceServer struct {
	payment.UnimplementedPaymentServiceServer
	DB           *gorm.DB
	Redis        *redis.Client
	Node         *snowflake.Node
	Proxy        proxy.ServiceProxy
	EventBus     events.EventBus
	OrderService order.OrderServiceClient // 订单服务客户端
}

// 信用卡相关常量
const (
	// 支付方式
	PaymentMethodCreditCard = "CREDIT_CARD"
	PaymentMethodAliPay     = "ALIPAY"
	PaymentMethodWeChatPay  = "WECHAT_PAY"

	// 卡类型
	CardTypeVisa       = "VISA"
	CardTypeMasterCard = "MASTERCARD"
	CardTypeUnionPay   = "UNIONPAY"
	CardTypeAmex       = "AMEX"
)

// 从信用卡号中获取后四位
func getLastFourDigits(cardNumber string) string {
	if len(cardNumber) <= 4 {
		return cardNumber
	}
	return cardNumber[len(cardNumber)-4:]
}

// 识别信用卡类型
func detectCardType(cardNumber string) string {
	if len(cardNumber) == 0 {
		return ""
	}

	// 简化的信用卡类型检测
	firstDigit := cardNumber[0]
	firstTwoDigits := ""
	if len(cardNumber) >= 2 {
		firstTwoDigits = cardNumber[0:2]
	}

	switch {
	case firstDigit == '4':
		return CardTypeVisa
	case firstTwoDigits >= "51" && firstTwoDigits <= "55":
		return CardTypeMasterCard
	case firstTwoDigits == "62":
		return CardTypeUnionPay
	case firstDigit == '3' && (cardNumber[1] == '4' || cardNumber[1] == '7'):
		return CardTypeAmex
	default:
		return "UNKNOWN"
	}
}
