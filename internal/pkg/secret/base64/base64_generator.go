package base64

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/JrMarcco/kuryr/internal/pkg/secret"
)

var _ secret.Generator = (*Generator)(nil)

type Generator struct{}

// Generate 生成业务密钥
// length: 生成密钥的字节长度（建议32字节，生成44字符的Base64字符串）
func (g *Generator) Generate(length int) (string, error) {
	if length < 16 {
		return "", fmt.Errorf("secret length should be at least 16 bytes")
	}

	// 使用加密安全的随机数生成器
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Base64编码，URL安全，无填充
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// GenerateWithPrefix 带前缀的密钥生成
func (g *Generator) GenerateWithPrefix(prefix string, length int) (string, error) {
	s, err := g.Generate(length)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s_%s", prefix, s), nil
}

func NewGenerator() *Generator {
	return &Generator{}
}
