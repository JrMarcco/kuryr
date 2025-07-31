package dao

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"time"

	"github.com/JrMarcco/kuryr/internal/domain"
	pkggorm "github.com/JrMarcco/kuryr/internal/pkg/gorm"
	"github.com/JrMarcco/kuryr/internal/search"
	"gorm.io/gorm"
)

const keySize = 32

type Provider struct {
	Id           uint64 `gorm:"column:id"`
	ProviderName string `gorm:"column:provider_name"`
	Channel      int32  `gorm:"column:channel"`

	Endpoint string `gorm:"column:endpoint"`
	RegionId string `gorm:"column:region_id"`

	AppId     string `gorm:"column:app_id"`
	ApiKey    string `gorm:"column:api_key"`
	ApiSecret string `gorm:"column:api_secret"`

	Weight     int32 `gorm:"column:weight"`
	QpsLimit   int32 `gorm:"column:qps_limit"`
	DailyLimit int32 `gorm:"column:daily_limit"`

	AuditCallbackUrl string `gorm:"column:audit_callback_url"`

	ActiveStatus string `gorm:"column:active_status"`
	CreatedAt    int64  `gorm:"column:created_at"`
	UpdatedAt    int64  `gorm:"column:updated_at"`
}

func (Provider) TableName() string {
	return "provider_info"
}

type ProviderDao interface {
	Save(ctx context.Context, provider Provider) error
	Delete(ctx context.Context, id uint64) error
	Update(ctx context.Context, provider Provider) error

	Search(ctx context.Context, criteria search.ProviderCriteria, param *pkggorm.PaginationParam) (*pkggorm.PaginationResult[Provider], error)
	FindById(ctx context.Context, id uint64) (Provider, error)
	FindByChannel(ctx context.Context, channel string) ([]Provider, error)
}

var _ ProviderDao = (*DefaultProviderDao)(nil)

type DefaultProviderDao struct {
	db         *gorm.DB
	encryptKey []byte
}

func (d *DefaultProviderDao) Save(ctx context.Context, provider Provider) error {
	now := time.Now().UnixMilli()
	provider.CreatedAt = now
	provider.UpdatedAt = now

	apiSecret := provider.ApiSecret
	encryptedSecret, err := d.encrypt(apiSecret)
	if err != nil {
		return err
	}

	//　db 保存加密后的密钥
	provider.ApiSecret = encryptedSecret
	return d.db.WithContext(ctx).Model(&Provider{}).Create(&provider).Error
}

// encrypt 使用 AES-GCM 加密
func (d *DefaultProviderDao) encrypt(plainText string) (string, error) {
	block, err := aes.NewCipher(d.encryptKey)
	if err != nil {

	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	cipherText := gcm.Seal(nonce, nonce, []byte(plainText), nil)
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

func (d *DefaultProviderDao) Delete(ctx context.Context, id uint64) error {
	return d.db.WithContext(ctx).Model(&Provider{}).
		Where("id = ?", id).
		Delete(&Provider{}).Error
}

func (d *DefaultProviderDao) Update(ctx context.Context, provider Provider) error {
	provider.UpdatedAt = time.Now().UnixMilli()

	values := map[string]any{
		"provider_name":      provider.ProviderName,
		"channel":            provider.Channel,
		"endpoint":           provider.Endpoint,
		"region_id":          provider.RegionId,
		"api_key":            provider.ApiKey,
		"weight":             provider.Weight,
		"qps_limit":          provider.QpsLimit,
		"daily_limit":        provider.DailyLimit,
		"audit_callback_url": provider.AuditCallbackUrl,
		"active_status":      provider.ActiveStatus,
		"updated_at":         provider.UpdatedAt,
	}

	if provider.ApiSecret != "" {
		encryptedSecret, err := d.encrypt(provider.ApiSecret)
		if err != nil {
			return err
		}
		values["api_secret"] = encryptedSecret
	}

	return d.db.WithContext(ctx).Model(&Provider{}).
		Where("id = ?", provider.Id).
		Updates(values).Error
}

func (d *DefaultProviderDao) Search(ctx context.Context, criteria search.ProviderCriteria, param *pkggorm.PaginationParam) (*pkggorm.PaginationResult[Provider], error) {
	var records []Provider

	query := d.db.WithContext(ctx).Model(&Provider{})
	if criteria.ProviderName != "" {
		query = query.Where("provider_name like ?", pkggorm.BuildLikePattern(criteria.ProviderName))
	}
	if criteria.Channel != 0 {
		query = query.Where("channel = ?", criteria.Channel)
	}
	return pkggorm.Pagination(query, param, records)
}

func (d *DefaultProviderDao) FindById(ctx context.Context, id uint64) (Provider, error) {
	var provider Provider
	err := d.db.WithContext(ctx).Model(&Provider{}).
		Where("id = ?", id).
		First(&provider).Error
	if err != nil {
		return Provider{}, err
	}

	if provider.ApiSecret != "" {
		decryptedSecret, err := d.decrypt(provider.ApiSecret)
		if err != nil {
			return Provider{}, err
		}
		provider.ApiSecret = decryptedSecret
	}
	return provider, nil
}

func (d *DefaultProviderDao) FindByChannel(ctx context.Context, channel string) ([]Provider, error) {
	var providers []Provider

	err := d.db.WithContext(ctx).Model(&Provider{}).
		Where("channel = ? and active_status = ?", channel, domain.ActiveStatusActive).
		Find(&providers).Error
	if err != nil {
		return nil, err
	}

	for i := range providers {
		if providers[i].ApiSecret != "" {
			decryptedSecret, err := d.decrypt(providers[i].ApiSecret)
			if err != nil {
				return nil, err
			}
			providers[i].ApiSecret = decryptedSecret
		}
	}
	return providers, nil
}

func (d *DefaultProviderDao) decrypt(encrypted string) (string, error) {
	cipherText, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(d.encryptKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	if len(cipherText) < gcm.NonceSize() {
		return "", errors.New("ciphertext too short")
	}

	nonce, cipherText := cipherText[:gcm.NonceSize()], cipherText[gcm.NonceSize():]

	plainText, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", err
	}
	return string(plainText), nil
}

func NewDefaultProviderDao(db *gorm.DB, encryptKey string) *DefaultProviderDao {
	// 确保 encrypt key 长度为 32 字节
	key := make([]byte, keySize)
	copy(key, encryptKey)

	return &DefaultProviderDao{
		db:         db,
		encryptKey: key,
	}
}
