package gorm

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
