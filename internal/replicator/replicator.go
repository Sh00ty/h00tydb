package replicator

import (
	"context"
	"errors"

	"gitlab.com/Sh00ty/hootydb/internal/kv"
)

type Replicator interface {
	Read(ctx context.Context, key kv.Key) (kv.Value, error)
	Write(ctx context.Context, key kv.Key, val any) error
}

type Resolver interface {
	GetRemoteAddrs(ctx context.Context) []string
	GetSelf(ctx context.Context) string
}

type Client interface {
	Read(ctx context.Context, key kv.Key) (val kv.Value, err error)
	Write(ctx context.Context, key kv.Key, val kv.Value) error
}

type Sequence interface {
	Inc() uint64
	Get() uint64
}

type ReplicatorError error

var (
	ErrReadPhase         ReplicatorError = errors.New("read phase")
	ErrWritePhase        ReplicatorError = errors.New("write phase")
	ErrQuorumNotGathered                 = errors.New("the request did not gather a quorum")
)
