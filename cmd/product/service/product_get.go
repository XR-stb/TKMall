package service

import (
	"context"

	"TKMall/build/proto_gen/product"
	"TKMall/cmd/product/model"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ProductCatalogServiceServer) GetProduct(ctx context.Context, req *product.GetProductReq) (*product.GetProductResp, error) {
	// 参数校验
	if req.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "商品ID不能为空")
	}

	// 尝试从缓存获取
	if cachedProduct, err := s.getCachedProduct(req.Id); err == nil {
		return &product.GetProductResp{Product: cachedProduct}, nil
	}

	// 数据库查询
	var productModel model.Product
	err := s.DB.Preload("Category").Where("id = ?", req.Id).First(&productModel).Error

	if err != nil {
		return nil, handleGetError(err)
	}

	// 转换proto格式
	protoProduct := convertToProtoProduct(&productModel)

	// 缓存结果
	s.cacheProduct(protoProduct)

	return &product.GetProductResp{
		Product: protoProduct,
	}, nil
}

func convertToProtoProduct(p *model.Product) *product.Product {
	protoProduct := &product.Product{
		Id:          uint32(p.ID),
		Name:        p.Name,
		Description: p.Description,
		Price:       float32(p.Price),
		Picture:     p.Images,
	}

	// 添加分类信息
	if p.Category.ID > 0 {
		protoProduct.Categories = []string{p.Category.Name}
	}

	return protoProduct
}
