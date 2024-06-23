package quorum

import (
	"context"
	"fmt"

	"gitlab.com/Sh00ty/hootydb/internal"
	"gitlab.com/Sh00ty/hootydb/internal/config"
	"gitlab.com/Sh00ty/hootydb/internal/kv"
	"gitlab.com/Sh00ty/hootydb/internal/quorum"
	"gitlab.com/Sh00ty/hootydb/internal/replicator"
	"gitlab.com/Sh00ty/hootydb/internal/versions"
)

// Что делать если сообщение не едет обратно

type Replicator struct {
	cfg           *config.Config
	seq           replicator.Sequence
	verGenerator  versions.VersionGenerator
	qreader       quorum.Qreader
	qwriter       quorum.QWriter
	replicas      map[string]replicator.Client
	replicasAddrs []string
	log           internal.Logger
}

func NewReplicator(
	cfg *config.Config,
	seq replicator.Sequence,
	replicas map[string]replicator.Client,
	log internal.Logger,
) *Replicator {
	replicasAddrs := make([]string, 0, len(replicas))
	for addr := range replicas {
		replicasAddrs = append(replicasAddrs, addr)
	}
	return &Replicator{
		cfg:           cfg,
		seq:           seq,
		replicas:      replicas,
		replicasAddrs: replicasAddrs,
		log:           log,
	}
}

func (r *Replicator) Read(ctx context.Context, key kv.Key) (kv.Value, error) {
readLoop:
	for i := 0; i < r.cfg.ReadRetry; i++ {
		read, bestVal, err := r.qreader.ReadPhase(ctx, key)
		if err != nil {
			return kv.Value{}, err
		}
		// fast path
		if bestVal.Ver.IsNull() {
			return kv.Value{}, kv.MakeNotFound(key)
		}

		r.log.Infof(ctx, "best version for %s -> %s", key, bestVal.Ver)

		replicasWithOldVersions := make([]string, 0, len(r.replicas))
		for _, addr := range r.replicasAddrs {
			val, exists := read[addr]
			if !exists {
				replicasWithOldVersions = append(replicasWithOldVersions, addr)
				continue
			}
			if val.Ver != bestVal.Ver {
				replicasWithOldVersions = append(replicasWithOldVersions, addr)
			}
		}

		writeRes, err := r.qwriter.WritePhase(ctx, key, bestVal, replicasWithOldVersions)
		if err != nil {
			return kv.Value{}, err
		}
		for _, wr := range writeRes {
			if wr.IsOldVer {
				r.log.Warnf(ctx, "Read: detected newer version of key while read from: %v", wr.Addr)
				continue readLoop
			}
		}
		return bestVal, nil
	}

	return kv.Value{}, fmt.Errorf("failed to read due to concurrent writes")
}

func (r *Replicator) Write(ctx context.Context, key kv.Key, val any) error {
writeLoop:
	for i := 0; i < r.cfg.WriteRetry; i++ {
		ver, err := r.verGenerator.GetNext(ctx, string(key))
		if err != nil {
			return err
		}

		writeRes, err := r.qwriter.WritePhase(ctx, key, kv.Value{
			Val: val,
			Ver: ver,
		}, r.replicasAddrs)
		if err != nil {
			return err
		}

		for _, wr := range writeRes {
			if wr.IsOldVer {
				r.log.Warnf(ctx, "Write: detected newer version of key while writing on: %v", wr.Addr)
				continue writeLoop
			}
		}
		return nil
	}

	return fmt.Errorf("failed to write due to concurrent writes")
}

// func (r *Replicator) generateVersion(ctx context.Context, key kv.Key) (versions.Version, error) {
// 	_, bestVal, err := r.qreader.ReadPhase(ctx, key)
// 	if err != nil {
// 		return nil, err
// 	}

// 	ver := bestVal.Ver
// 	ver.SeqNum++
// 	ver.InternalSeqNum = r.seq.Inc()
// 	ver.NodeID = r.cfg.NodeID
// 	return ver, nil
// }
