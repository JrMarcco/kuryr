package jwt

import (
	"context"
	"fmt"

	kuryrapi "github.com/JrMarcco/kuryr-api/api/go"
	"github.com/JrMarcco/kuryr/internal/errs"
)

const (
	paramBizId  = "biz_id"
	paramBizKey = "biz_key"
)

func ContextBizId(ctx context.Context) (uint64, error) {
	val := ctx.Value(kuryrapi.ContextKeyBizId{})
	if val == nil {
		return 0, fmt.Errorf("%w: biz id is nil", errs.ErrInvalidParam)
	}
	bizId, ok := val.(uint64)
	if !ok {
		return 0, fmt.Errorf("%w: biz id is not uint64", errs.ErrInvalidParam)
	}
	return bizId, nil
}
