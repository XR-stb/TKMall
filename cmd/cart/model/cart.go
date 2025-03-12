package model

import (
	"TKMall/common/model"
	"time"

	"gorm.io/gorm"
)

// Cart 购物车模型
type Cart struct {
	model.BaseModel
	UserID    int64     `gorm:"index;not null"` // 用户ID
	UpdatedAt time.Time // 最后更新时间
}

// CartItem 购物车项模型
type CartItem struct {
	model.BaseModel
	CartID    uint    `gorm:"index;not null"`              // 购物车ID
	ProductID uint    `gorm:"index;not null"`              // 商品ID
	Quantity  int     `gorm:"not null"`                    // 商品数量
	Price     float64 `gorm:"type:decimal(10,2);not null"` // 商品单价（加入时的价格）
}

// 初始化数据库表
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&Cart{},
		&CartItem{},
	)
}
