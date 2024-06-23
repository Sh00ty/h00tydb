package httpsrv

import (
	"gitlab.com/Sh00ty/hootydb/internal"
	"gitlab.com/Sh00ty/hootydb/internal/kv"
	"gitlab.com/Sh00ty/hootydb/internal/sharder"
)

type Srv struct {
	sharder sharder.Sharder
	kv      kv.KV
	log     internal.Logger
}

func NewSrv(sharder sharder.Sharder, kv kv.KV, log internal.Logger) *Srv {
	return &Srv{
		sharder: sharder,
		kv:      kv,
		log:     log,
	}
}
