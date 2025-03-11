package model

import (
	"TKMall/common/model"
	"time"

	"gorm.io/gorm"
)

type Product struct {
	model.BaseModel
	Name        string          `gorm:"type:varchar(100);not null;index:idx_search,priority:1"`
	Description string          `gorm:"type:text;index:idx_search,priority:2,length:255"`
	Price       float64         `gorm:"type:decimal(10,2);not null;index"`
	Stock       int             `gorm:"type:int unsigned;not null;default:0"`
	CategoryID  uint            `gorm:"index"`
	Category    ProductCategory `gorm:"foreignKey:CategoryID"`
	IsPublished bool            `gorm:"default:false"`
	PublishedAt time.Time
	Images      string `gorm:"type:text"` // JSON数组存储图片路径
}

type ProductCategory struct {
	model.BaseModel
	Name        string `gorm:"type:varchar(50);uniqueIndex;not null"`
	Description string `gorm:"type:text"`
	SortOrder   int    `gorm:"type:int;default:0"`
}

type ProductSKU struct {
	model.BaseModel
	ProductID uint    `gorm:"index;not null"`
	SKU       string  `gorm:"type:varchar(50);uniqueIndex;not null"`
	Price     float64 `gorm:"type:decimal(10,2);not null"`
	Stock     int     `gorm:"type:int unsigned;not null;default:0"`
	Specs     string  `gorm:"type:json"` // 规格参数，JSON格式
}

// 初始化数据库表
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&Product{},
		&ProductCategory{},
		&ProductSKU{},
	)
}
