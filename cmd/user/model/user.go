package model

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type User struct {
	ID          int64 `gorm:"primaryKey"`
	Email       string `gorm:"unique;not null"`
	Password    string `gorm:"not null"`
	Username    string `gorm:"not null"`
	FirstName   string
	LastName    string
	PhoneNumber string
	Address     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type UserSettings struct {
	UserID              int64     `gorm:"primaryKey"`
	Theme               string
	NotificationEnabled bool
	Language            string
	CreatedAt           time.Time
}

// InitGORM 初始化 GORM 数据库连接
func InitGORM() (*gorm.DB, error) {
	dsn := viper.GetString("mysql.dsn")

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open GORM connection: %v", err)
	}

	// 自动迁移数据库
	err = db.AutoMigrate(&User{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %v", err)
	}

	return db, nil
}
