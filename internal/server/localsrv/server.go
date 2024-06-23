package localsrv

import (
	"gitlab.com/Sh00ty/hootydb/internal/kv"
	"gitlab.com/Sh00ty/hootydb/internal/transport/transport"
	"gitlab.com/Sh00ty/hootydb/internal/versions"
)

type Srv struct {
	transport.UnsafeReplicatorServiceServer
	verBuilder   versions.VersionBuilder
	storage      kv.KV
	faultPercent int
}

func New(store kv.KV, verBuilder versions.VersionBuilder, faultPercent int) transport.ReplicatorServiceServer {
	return &Srv{
		storage:      store,
		verBuilder:   verBuilder,
		faultPercent: faultPercent,
	}
}
