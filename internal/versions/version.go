package versions

import (
	"context"
)

type Version interface {
	IsBigger(ver Version) bool
	IsNull() bool
	String() string
}

type VersionGenerator interface {
	GetNext(ctx context.Context, key string) (Version, error)
}

type VersionBuilder interface {
	FromString(string) (Version, error)
}
