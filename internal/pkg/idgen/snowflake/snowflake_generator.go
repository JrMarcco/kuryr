package snowflake

import (
	"sync/atomic"
	"time"

	"github.com/JrMarcco/kuryr/internal/pkg/hash"
	"github.com/JrMarcco/kuryr/internal/pkg/idgen"
)

const (
	timestampBits = 41
	hashBits      = 10
	sequenceBits  = 12

	hashShift      = sequenceBits
	timestampShift = hashShift + hashBits

	sequenceMask  = (uint64(1) << sequenceBits) - 1
	hashMask      = (uint64(1) << hashBits) - 1
	timestampMask = (uint64(1) << timestampBits) - 1

	// epoch time 设置成 2025-01-01 00:00:00
	epochMillis   = uint64(1735689600000) // milliseconds of 2025-01-01 00:00:00
	number1000    = uint64(1000)
	number1000000 = uint64(1000000)
)

var _ idgen.Generator = (*Generator)(nil)

// Generator 自定义的雪花算法 id 生成器。
// 其中机器码的 10 位替换为由业务 id 和业务 key 计算而来的 hash 值。
// 这样通过 id 能解析出业务 id 和业务 key 从而获得分库分表信息。
type Generator struct {
	sequence uint64
	lastTime uint64 // 上一次生成 id 的时间

	epoch time.Time
}

// NextId 生成 id。
//
// id 组成信息:
// ├── 41 位时间戳，几点时间为 2025-01-01 00:00:00。
// ├── 10 位 hash 值，由业务 id（bizId）和业务 key（bizKey）一起计算出的 hash 值。
// ├── 12 位自增序列。
func (g *Generator) NextId(bizId uint64, bizKey string) uint64 {
	timestamp := uint64(time.Now().UnixMilli()) - epochMillis
	hashVal := hash.HashUint64(bizId, bizKey)
	seq := atomic.AddUint64(&g.sequence, 1) - 1

	return (timestamp&timestampMask)<<timestampShift | (hashVal&hashMask)<<hashShift | (seq & sequenceMask)
}

func NewGenerator() *Generator {
	return &Generator{
		sequence: 0,
		lastTime: 0,
		epoch:    time.Unix(int64(epochMillis/number1000), int64((epochMillis%number1000)*number1000000)),
	}
}

func ExtractHash(id uint64) uint64 {
	return (id >> hashShift) & hashMask
}

func ExtractSequence(id uint64) uint64 {
	return id & sequenceMask
}

func ExtractTimestamp(id uint64) time.Time {
	timestamp := (id >> timestampShift) & timestampMask
	return time.Unix(0, int64((timestamp+epochMillis)*uint64(time.Millisecond)))
}
