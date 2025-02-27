package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"TKMall/build/proto_gen/product"
	"TKMall/cmd/product/model"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

func (s *ProductCatalogServiceServer) SearchProducts(ctx context.Context, req *product.SearchProductsReq) (*product.SearchProductsResp, error) {
	// 参数校验
	if strings.TrimSpace(req.Query) == "" {
		return nil, status.Error(codes.InvalidArgument, "搜索关键词不能为空")
	}

	// 构建搜索查询
	searchTerm := fmt.Sprintf("%%%s%%", strings.ToLower(req.Query))
	query := s.DB.Model(&model.Product{}).
		Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", searchTerm, searchTerm)

	// 执行查询
	var products []model.Product
	if err := query.Find(&products).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &product.SearchProductsResp{Results: []*product.Product{}}, nil
		}
		return nil, status.Errorf(codes.Internal, "搜索失败: %v", err)
	}

	// 转换结果
	var results []*product.Product
	for _, p := range products {
		results = append(results, &product.Product{
			Id:          uint32(p.ID),
			Name:        p.Name,
			Description: p.Description,
			Price:       float32(p.Price),
			Picture:     p.Images,
		})
	}

	return &product.SearchProductsResp{
		Results: results,
	}, nil
}
