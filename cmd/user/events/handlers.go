package events

import (
	"context"
	"fmt"
	"time"

	"TKMall/cmd/user/model"
	"TKMall/common/events"
	"TKMall/common/log"
)

// 用户注册事件的payload结构
type UserRegisteredPayload struct {
	UserID    int64     `json:"user_id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// 初始化事件处理器
func InitEventHandlers(eventBus events.EventBus) error {
	// 注册用户注册事件处理器
	eventBus.Subscribe(events.UserRegistered, handleUserRegistered)
	return nil
}

// 处理用户注册事件
func handleUserRegistered(ctx context.Context, event events.Event) error {
	payload, ok := event.Payload.(UserRegisteredPayload)
	if !ok {
		return fmt.Errorf("invalid payload type")
	}

	// 并行处理各个任务
	errChan := make(chan error, 3)

	// 1. 发送欢迎邮件
	go func() {
		errChan <- sendWelcomeEmail(ctx, payload.Email)
	}()

	// 2. 初始化用户设置
	go func() {
		errChan <- initializeUserSettings(ctx, payload.UserID)
	}()

	// 3. 发送管理员通知
	go func() {
		errChan <- notifyAdmin(ctx, payload)
	}()

	// 收集错误
	var errors []error
	for i := 0; i < 3; i++ {
		if err := <-errChan; err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("some handlers failed: %v", errors)
	}
	return nil
}

// 发送欢迎邮件
func sendWelcomeEmail(ctx context.Context, email string) error {
	// TODO: 实现邮件发送逻辑
	log.Infof("Sending welcome email to: %s", email)
	return nil
}

// 初始化用户设置
func initializeUserSettings(ctx context.Context, userID int64) error {
	_ = &model.UserSettings{
		UserID:              userID,
		Theme:               "default",
		NotificationEnabled: true,
		Language:            "zh_CN",
		CreatedAt:           time.Now(),
	}

	// TODO: 保存到数据库
	log.Infof("Initialized settings for user: %d", userID)
	return nil
}

// 通知管理员
func notifyAdmin(ctx context.Context, payload UserRegisteredPayload) error {
	// TODO: 实现管理员通知逻辑
	log.Infof("New user registered: %v", payload)
	return nil
}
