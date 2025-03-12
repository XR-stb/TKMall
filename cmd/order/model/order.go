package model

import (
	"TKMall/common/model"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// 地址信息
type Address struct {
	StreetAddress string `json:"street_address"`
	City          string `json:"city"`
	State         string `json:"state"`
	Country       string `json:"country"`
	ZipCode       int32  `json:"zip_code"`
}

// 实现SQL驱动接口，使其可以被GORM作为JSON存储
func (a Address) Value() (driver.Value, error) {
	return json.Marshal(a)
}

// 实现Scanner接口，用于从数据库中读取JSON
func (a *Address) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("类型断言为[]byte失败")
	}
	return json.Unmarshal(bytes, &a)
}

// 订单项
type OrderItem struct {
	model.BaseModel
	OrderID    string  `gorm:"type:varchar(50);index;not null"`
	ProductID  uint    `gorm:"index;not null"`
	Quantity   int     `gorm:"not null"`
	Price      float64 `gorm:"type:decimal(10,2);not null"` // 下单时的商品单价
	TotalPrice float64 `gorm:"type:decimal(10,2);not null"` // 商品总价
}

// 订单状态
type OrderStatus string

const (
	OrderStatusCreated   OrderStatus = "CREATED"   // 订单创建
	OrderStatusPaid      OrderStatus = "PAID"      // 已支付
	OrderStatusShipped   OrderStatus = "SHIPPED"   // 已发货
	OrderStatusDelivered OrderStatus = "DELIVERED" // 已送达
	OrderStatusCancelled OrderStatus = "CANCELLED" // 已取消
	OrderStatusReturned  OrderStatus = "RETURNED"  // 已退货
)

// 订单
type Order struct {
	model.BaseModel
	OrderID       string      `gorm:"type:varchar(50);uniqueIndex;not null"`       // 订单号
	UserID        int64       `gorm:"index;not null"`                              // 用户ID
	Status        OrderStatus `gorm:"type:varchar(20);not null;default:'CREATED'"` // 订单状态
	TotalAmount   float64     `gorm:"type:decimal(10,2);not null"`                 // 订单总金额
	Address       Address     `gorm:"type:json"`                                   // 收货地址
	Email         string      `gorm:"type:varchar(100)"`                           // 用户邮箱
	UserCurrency  string      `gorm:"type:varchar(10);default:'CNY'"`              // 用户货币类型
	PaymentID     string      `gorm:"type:varchar(50);index"`                      // 支付ID
	TransactionID string      `gorm:"type:varchar(100)"`                           // 交易ID
	PaidAt        *time.Time  `gorm:"index"`                                       // 支付时间
	ShippedAt     *time.Time  // 发货时间
	DeliveredAt   *time.Time  // 送达时间
	CancelledAt   *time.Time  // 取消时间
}

// 初始化数据库表
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&Order{},
		&OrderItem{},
	)
}
