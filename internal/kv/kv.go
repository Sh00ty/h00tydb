package kv

import (
	"context"
	"errors"
	"fmt"

	"gitlab.com/Sh00ty/hootydb/internal/versions"
)

type Key string

type Value struct {
	Val any
	Ver versions.Version
}

type ErrKind int

const (
	Unknown ErrKind = iota
	NotFound
	OldVer
)

type KVError struct {
	Kind  ErrKind
	Extra []any
	Key   Key
}

func (e *KVError) Error() string {
	switch e.Kind {
	case NotFound:
		return fmt.Sprintf("not found key: %s", e.Key)
	case OldVer:
		return fmt.Sprintf("can't modify key %s: too old version %v, new is %v", e.Key, e.Extra[0], e.Extra[1])
	}
	if e.Extra != nil {
		return fmt.Sprintf("unknown error for key: %s and extra: %v", e.Key, e.Extra)
	}
	return fmt.Sprintf("got unknown error for key: %s", e.Key)
}

func MakeOldVerError(key Key, old, new versions.Version) *KVError {
	return &KVError{
		Key:   key,
		Extra: []any{old, new},
		Kind:  OldVer,
	}
}

func MakeNotFound(key Key) *KVError {
	return &KVError{
		Key:  key,
		Kind: NotFound,
	}
}

func MakeUnknown(key Key, extra ...any) *KVError {
	return &KVError{
		Key:   key,
		Extra: extra,
		Kind:  Unknown,
	}
}

func IsNotFound(err error) bool {
	e := &KVError{}
	if !errors.As(err, &e) {
		return false
	}
	return e.Kind == NotFound
}

func IsOldVer(err error) bool {
	e := &KVError{}
	if !errors.As(err, &e) {
		return false
	}
	return e.Kind == OldVer
}

type KV interface {
	Get(ctx context.Context, k Key) (Value, error)
	Set(ctx context.Context, k Key, v Value) error
}
