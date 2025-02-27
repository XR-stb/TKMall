package service

import (
	"TKMall/build/proto_gen/product"
	"context"
	"encoding/json"
	"fmt"
	"time"
)

const (
	productListCacheKey = "product_list:%s:%s:%f:%f:%s:%d:%d"
	cacheExpiration     = 5 * time.Minute
	productCacheKey     = "product:%d"
)

// func (s *ProductService) getCachedProducts(req *product.ListProductsReq) ([]*product.Product, int32, error) {
// 	_ = fmt.Sprintf(productListCacheKey,
// 		req.CategoryName,
// 		req.Page,
// 		req.PageSize,
// 	)

// 	// 从Redis获取缓存
// 	// ...
// 	return nil, 0, nil
// }

// func (s *ProductService) cacheProducts(req *product.ListProductsReq, products []*product.Product, total int32) {
// 	// 存储到Redis
// 	// ...
// }

func (s *ProductCatalogServiceServer) getCachedProduct(productID uint32) (*product.Product, error) {
	cacheKey := fmt.Sprintf(productCacheKey, productID)

	// 从Redis获取缓存
	cachedData, err := s.Redis.Get(context.Background(), cacheKey).Result()
	if err == nil {
		var cachedProduct product.Product
		if err := json.Unmarshal([]byte(cachedData), &cachedProduct); err == nil {
			return &cachedProduct, nil
		}
	}

	return nil, err
}

func (s *ProductCatalogServiceServer) cacheProduct(product *product.Product) {
	cacheKey := fmt.Sprintf(productCacheKey, product.Id)
	productData, _ := json.Marshal(product)
	s.Redis.Set(context.Background(), cacheKey, productData, 30*time.Minute)
}
