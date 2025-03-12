package service

import (
	"context"
	"time"

	"TKMall/build/proto_gen/order"
	"TKMall/cmd/order/model"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MarkOrderPaid 标记订单为已支付
func (s *OrderServiceServer) MarkOrderPaid(ctx context.Context, req *order.MarkOrderPaidReq) (*order.MarkOrderPaidResp, error) {
	// 参数校验
	if req.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "用户ID不能为空")
	}
	if req.OrderId == "" {
		return nil, status.Error(codes.InvalidArgument, "订单ID不能为空")
	}

	// 查询订单是否存在
	var orderInfo model.Order
	if err := s.DB.Where("order_id = ? AND user_id = ?", req.OrderId, req.UserId).First(&orderInfo).Error; err != nil {
		return nil, status.Errorf(codes.NotFound, "订单不存在: %v", err)
	}

	// 检查订单状态
	if orderInfo.Status != model.OrderStatusCreated {
		return nil, status.Errorf(codes.FailedPrecondition, "订单状态不正确，当前状态: %s", orderInfo.Status)
	}

	// 获取当前时间
	now := time.Now()

	// 更新订单状态
	updates := map[string]interface{}{
		"status":  model.OrderStatusPaid,
		"paid_at": now,
	}

	if err := s.DB.Model(&orderInfo).Updates(updates).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "更新订单状态失败: %v", err)
	}

	// 发布订单支付事件
	if s.EventBus != nil {
		// 这里可以发布订单支付事件，供其他服务订阅处理
		// s.EventBus.Publish("order.paid", event)
	}

	return &order.MarkOrderPaidResp{}, nil
}
