package service

import (
	"time"
)

const (
	productListCacheKey = "product_list:%s:%s:%f:%f:%s:%d:%d"
	cacheExpiration     = 5 * time.Minute
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
