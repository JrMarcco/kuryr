package hash

import (
	"strconv"

	"github.com/cespare/xxhash/v2"
)

func HashUint64(bizId uint64, bizKey string) uint64 {
	return xxhash.Sum64String(strconv.FormatUint(bizId, 10) + ":" + bizKey)
}
