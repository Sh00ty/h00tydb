package localclnt

import (
	"context"
	"fmt"

	"gitlab.com/Sh00ty/hootydb/internal/kv"
	"gitlab.com/Sh00ty/hootydb/internal/utils/random"
)

type Client struct {
	storage         kv.KV
	faultPercentage int
}

func New(stor kv.KV, faultPercentage int) *Client {
	return &Client{
		storage:         stor,
		faultPercentage: faultPercentage,
	}
}

func (c *Client) Read(ctx context.Context, key kv.Key) (val kv.Value, err error) {
	if random.Random(100) <= c.faultPercentage {
		return kv.Value{}, fmt.Errorf("db fault injection")
	}

	return c.storage.Get(ctx, key)
}
func (c *Client) Write(ctx context.Context, key kv.Key, val kv.Value) error {
	if random.Random(100) <= c.faultPercentage {
		return fmt.Errorf("db fault injection")
	}

	return c.storage.Set(ctx, key, val)
}
