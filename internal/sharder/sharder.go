package sharder

import (
	"context"

	"gitlab.com/Sh00ty/hootydb/internal/kv"
)

// TODO: hinted handoffs (реплицируем не в свой рендж)
// если не удалось собрать кворум по своему ренджу

type Sharder interface {
	Read(ctx context.Context, key kv.Key) (kv.Value, error)
	Write(ctx context.Context, key kv.Key, val any) error
}

type ReplicaSet interface {
	Read(ctx context.Context, key kv.Key) (kv.Value, error)
	Write(ctx context.Context, key kv.Key, val any) error
}
