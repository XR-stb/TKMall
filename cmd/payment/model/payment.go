package model

import (
	"TKMall/common/model"
	"time"

	"gorm.io/gorm"
)

// 支付状态
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "PENDING"   // 等待支付
	PaymentStatusCompleted PaymentStatus = "COMPLETED" // 支付完成
	PaymentStatusFailed    PaymentStatus = "FAILED"    // 支付失败
	PaymentStatusRefunded  PaymentStatus = "REFUNDED"  // 已退款
	PaymentStatusCancelled PaymentStatus = "CANCELLED" // 已取消
)

// 信用卡信息（脱敏存储）
type CreditCard struct {
	model.BaseModel
	UserID             int64  `gorm:"index;not null"`            // 用户ID
	LastFourDigits     string `gorm:"type:varchar(4);not null"`  // 卡号后四位
	ExpirationMonth    int    `gorm:"not null"`                  // 过期月份
	ExpirationYear     int    `gorm:"not null"`                  // 过期年份
	CardType           string `gorm:"type:varchar(20);not null"` // 卡类型（Visa, Mastercard等）
	BillingAddressHash string `gorm:"type:varchar(64);not null"` // 账单地址哈希（用于验证）
}

// 交易记录
type Transaction struct {
	model.BaseModel
	TransactionID      string        `gorm:"type:varchar(100);uniqueIndex;not null"` // 交易ID
	OrderID            string        `gorm:"type:varchar(50);index;not null"`        // 关联的订单ID
	UserID             int64         `gorm:"index;not null"`                         // 用户ID
	Amount             float64       `gorm:"type:decimal(10,2);not null"`            // 交易金额
	Currency           string        `gorm:"type:varchar(10);default:'CNY'"`         // 货币类型
	Status             PaymentStatus `gorm:"type:varchar(20);index;not null"`        // 交易状态
	PaymentMethod      string        `gorm:"type:varchar(20);not null"`              // 支付方式
	CreditCardID       uint          `gorm:"index"`                                  // 信用卡ID
	TransactionTime    *time.Time    `gorm:"index"`                                  // 交易时间
	LastFourDigits     string        `gorm:"type:varchar(4)"`                        // 卡号后四位（冗余存储）
	GatewayResponseRaw string        `gorm:"type:text"`                              // 支付网关原始响应
	ErrorCode          string        `gorm:"type:varchar(50)"`                       // 错误代码
	ErrorMessage       string        `gorm:"type:varchar(255)"`                      // 错误信息
}

// 初始化数据库表
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&CreditCard{},
		&Transaction{},
	)
}
