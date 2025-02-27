package service

import (
	"context"

	"TKMall/build/proto_gen/product"
	"TKMall/cmd/product/model"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ProductCatalogServiceServer) ListProducts(ctx context.Context, req *product.ListProductsReq) (*product.ListProductsResp, error) {
	// 构建查询条件
	query := s.DB.Model(&model.Product{})

	if req.CategoryName != "" {
		query = query.Joins("JOIN product_categories ON products.category_id = product_categories.id").
			Where("product_categories.name = ?", req.CategoryName)
	}

	// 分页处理
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "获取总数失败: %v", err)
	}

	pageSize := int(req.PageSize)
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}
	currentPage := int(req.Page)
	if currentPage < 1 {
		currentPage = 1
	}
	offset := (currentPage - 1) * pageSize

	var products []model.Product
	if err := query.Offset(offset).Limit(pageSize).Find(&products).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "查询失败: %v", err)
	}

	// 转换为proto格式
	var protoProducts []*product.Product
	for _, p := range products {
		protoProducts = append(protoProducts, &product.Product{
			Id:          uint32(p.ID),
			Name:        p.Name,
			Description: p.Description,
			Price:       float32(p.Price),
		})
	}

	return &product.ListProductsResp{
		Products: protoProducts,
	}, nil
}
