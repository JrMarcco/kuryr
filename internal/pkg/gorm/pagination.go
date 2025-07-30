package gorm

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// BuildLikePattern 构建LIKE查询模式
func BuildLikePattern(s string) string {
	if s == "" {
		return ""
	}
	return "%" + sanitizeStringForLike(s) + "%"
}

// sanitizeStringForLike 清理字符串用于 LIKE 查询
func sanitizeStringForLike(s string) string {
	// 移除或转义特殊字符，防止SQL注入
	s = strings.ReplaceAll(s, "%", "\\%")
	s = strings.ReplaceAll(s, "_", "\\_")
	s = strings.ReplaceAll(s, "'", "\\'")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return strings.TrimSpace(s)
}

// Pagination 分页查询
func Pagination[T any](db *gorm.DB, param *PaginationParam, records []T) (*PaginationResult[T], error) {
	var total int64
	// 克隆查询以避免影响原查询
	countDB := db.Session(&gorm.Session{})
	if err := countDB.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("[kuryr] failed to count records: %w", err)
	}

	if total == 0 {
		var empty []T
		return NewPaginationResult(empty, 0), nil
	}

	if err := db.Offset(param.Offset).Limit(param.Limit).Find(&records).Error; err != nil {
		return nil, fmt.Errorf("[kuryr] failed to query records: %w", err)
	}
	return NewPaginationResult(records, total), nil
}
