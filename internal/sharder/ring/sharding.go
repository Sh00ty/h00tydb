package ring

import (
	"context"
	"slices"

	"gitlab.com/Sh00ty/hootydb/internal/kv"
	"gitlab.com/Sh00ty/hootydb/internal/sharder"
	"golang.org/x/exp/slices"
)

type Hasher interface {
	Hash(k kv.Key) int
}

type ringPoint struct {
	r     sharder.ReplicaSet
	start int
	end   int
}

// хочу получать датацентры и на их основе зная сколько там машин научиться
// делать реплика сеты, они уже автоматически будут реплицироваться)))
type Ring struct {
	h      Hasher
	points []ringPoint
}

func (r *Ring) findPoint(key kv.Key) ringPoint {
	hash := r.h.Hash(key)
	idx, _ := slices.BinarySearchFunc(r.points, hash, func(r ringPoint, t int) int {
		if r.start < t {
			return -1
		}
		if r.start <= t && r.end < t {
			return 0
		}
		return 1
	})

}

type Sharder struct {
}

func (s *Sharder) Read(ctx context.Context, key kv.Key) (kv.Value, error)
func (s *Sharder) Write(ctx context.Context, key kv.Key, val any) error
