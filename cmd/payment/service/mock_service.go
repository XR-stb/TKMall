// 仅用于测试的模拟代码
//go:build test
// +build test

package service

import (
	"TKMall/build/proto_gen/payment"
	"context"

	"github.com/stretchr/testify/mock"
)

// 节点生成器接口，用于测试
type NodeGenerator interface {
	Generate() ID
}

// ID接口
type ID interface {
	String() string
}

// MockID实现ID接口
type MockID struct {
	id string
}

func (m MockID) String() string {
	return m.id
}

// 模拟节点生成器
type MockNode struct {
	mock.Mock
}

func (m *MockNode) Generate() ID {
	args := m.Called()
	return args.Get(0).(ID)
}

// 测试服务实现
type MockPaymentServer struct {
	mock.Mock
}

func (m *MockPaymentServer) Charge(ctx context.Context, req *payment.ChargeReq) (*payment.ChargeResp, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*payment.ChargeResp), args.Error(1)
}

// 测试辅助函数
func isValidCreditCard(cardNum string) bool {
	length := len(cardNum)
	return length >= 13 && length <= 19
}

func detectCardType(cardNum string) string {
	if len(cardNum) == 0 {
		return "Unknown"
	}

	firstDigit := cardNum[0]
	firstTwoDigits := ""
	if len(cardNum) >= 2 {
		firstTwoDigits = cardNum[0:2]
	}

	if firstDigit == '4' {
		return "Visa"
	} else if firstDigit == '5' {
		return "MasterCard"
	} else if firstDigit == '3' && len(cardNum) >= 2 && (cardNum[1] == '4' || cardNum[1] == '7') {
		return "AmericanExpress"
	} else if firstTwoDigits == "60" {
		return "Discover"
	}

	return "Unknown"
}

func getLastFourDigits(cardNum string) string {
	length := len(cardNum)
	if length == 0 {
		return ""
	}
	if length <= 4 {
		return cardNum
	}
	return cardNum[length-4:]
}
