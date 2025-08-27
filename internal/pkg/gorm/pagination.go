package gorm

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// PaginationParam 分页参数
type PaginationParam struct {
	Offset int `json:"offset" form:"offset"`
	Limit  int `json:"limit" form:"limit"`
}

// PaginationResult 分页查询结果
type PaginationResult[T any] struct {
	Records []T   `json:"records"`
	Total   int64 `json:"total"`
}

func NewPaginationResult[T any](records []T, total int64) *PaginationResult[T] {
	return &PaginationResult[T]{
		Records: records,
		Total:   total,
	}
}

// BuildLikePattern 构建LIKE查询模式
func BuildLikePattern(s string) string {
	if s == "" {
		return ""
	}

	// 移除或转义特殊字符，防止SQL注入
	s = strings.ReplaceAll(s, "%", "\\%")
	s = strings.ReplaceAll(s, "_", "\\_")
	s = strings.ReplaceAll(s, "'", "\\'")
	s = strings.ReplaceAll(s, "\"", "\\\"")

	return fmt.Sprintf("%%%s%%", strings.TrimSpace(s))
}

// Pagination 分页查询
func Pagination[T any](db *gorm.DB, param *PaginationParam, records []T) (*PaginationResult[T], error) {
	if param == nil {
		// 参数为空时使用默认值
		param = &PaginationParam{
			Offset: 0,
			Limit:  10,
		}
	}

	var total int64
	// 克隆查询以避免影响原查询
	countDB := db.Session(&gorm.Session{})
	if err := countDB.Count(&total).Error; err != nil {
		return nil, err
	}

	if total == 0 {
		var empty []T
		return NewPaginationResult(empty, 0), nil
	}

	if err := db.Offset(param.Offset).Limit(param.Limit).Find(&records).Error; err != nil {
		return nil, err
	}
	return NewPaginationResult(records, total), nil
}
