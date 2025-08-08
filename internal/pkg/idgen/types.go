package idgen

// Generator id 生成器，生成依据为 bizId 和 bizKey。
type Generator interface {
	NextId(bizId uint64, bizKey string) uint64
}
