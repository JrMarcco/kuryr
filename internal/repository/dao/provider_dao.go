package dao

import "gorm.io/gorm"

type Provider struct {
	Id           uint64 `gorm:"column:id"`
	ProviderName string `gorm:"column:provider_name"`
	Channel      string `gorm:"column:channel"`

	Endpoint string `gorm:"column:endpoint"`
	RegionId string `gorm:"column:region_id"`

	AppId     string `gorm:"column:app_id"`
	ApiKey    string `gorm:"column:api_key"`
	ApiSecret string `gorm:"column:api_secret"`

	Weight     int `gorm:"column:weight"`
	QpsLimit   int `gorm:"column:qps_limit"`
	DailyLimit int `gorm:"column:daily_limit"`

	AuditCallbackUrl string `gorm:"column:audit_callback_url"`

	ActiveStatus string `gorm:"column:active_status"`
	CreatedAt    int64  `gorm:"column:created_at"`
	UpdatedAt    int64  `gorm:"column:updated_at"`
}

func (Provider) TableName() string {
	return "provider_info"
}

type ProviderDao interface{}

var _ ProviderDao = (*DefaultProviderDao)(nil)

type DefaultProviderDao struct {
	db *gorm.DB
}

func NewDefaultProviderDao(db *gorm.DB) *DefaultProviderDao {
	return &DefaultProviderDao{db: db}
}
