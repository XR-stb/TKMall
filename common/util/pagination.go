package util

import (
	"math"
)

func CalculatePagination(total int, page int, pageSize int) (currentPage int, totalPages int) {
	if pageSize <= 0 {
		pageSize = 10
	}
	totalPages = int(math.Ceil(float64(total) / float64(pageSize)))
	if page < 1 {
		page = 1
	} else if page > totalPages && totalPages > 0 {
		page = totalPages
	}
	return page, totalPages
} 