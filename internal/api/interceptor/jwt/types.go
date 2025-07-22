package jwt

import (
	"context"

	kuryrapi "github.com/JrMarcco/kuryr-api/api"
	"github.com/JrMarcco/kuryr/internal/errs"
)

const (
	paramBizId  = "biz_id"
	paramBizKey = "biz_key"
)

func ContextBizId(ctx context.Context) (uint64, error) {
	val := ctx.Value(kuryrapi.ContextKeyBizId{})
	if val == nil {
		return 0, errs.ErrInvalidBizId
	}
	bizId, ok := val.(uint64)
	if !ok {
		return 0, errs.ErrInvalidBizId
	}
	return bizId, nil
}
