package xxhash

import (
	"github.com/cespare/xxhash/v2"
	"gitlab.com/Sh00ty/hootydb/internal/kv"
)

type Hasher struct {
}

func NewHasher() *Hasher {
	return &Hasher{}
}

func (h *Hasher) Hash(k kv.Key) int {
	hash, err := xxhash.New().WriteString(string(k))
	if err != nil {
		panic(err)
	}
	return hash
}
