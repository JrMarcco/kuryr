package gorm

import "strings"

// IsUniqueConstraintError 判断是否为唯一索引冲突。
func IsUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())

	// postgresql 唯一键冲突错误关键词
	postgresKeywords := []string{
		"unique constraint",
		"duplicate key",
		"violates unique constraint",
	}

	// MySQL 唯一键冲突错误关键词
	mysqlKeywords := []string{
		"duplicate entry",
		"unique constraint",
	}

	keywords := append(postgresKeywords, mysqlKeywords...)
	for _, keyword := range keywords {
		if strings.Contains(errStr, keyword) {
			return true
		}
	}
	return false
}
