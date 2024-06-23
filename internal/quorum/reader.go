package quorum

import (
	"context"
	"fmt"

	"gitlab.com/Sh00ty/hootydb/internal"
	"gitlab.com/Sh00ty/hootydb/internal/config"
	"gitlab.com/Sh00ty/hootydb/internal/kv"
	"gitlab.com/Sh00ty/hootydb/internal/replicator"
)

type Qreader struct {
	cfg      config.Config
	log      internal.Logger
	replicas map[string]replicator.Client
}

func NewQreader(cfg config.Config, replicas map[string]replicator.Client, log internal.Logger) *Qreader {
	return &Qreader{
		cfg:      cfg,
		log:      log,
		replicas: replicas,
	}
}

type qReadRes struct {
	err  error
	addr string
	val  kv.Value
}

func (q *Qreader) ReadPhase(ctx context.Context, key kv.Key) (map[string]kv.Value, kv.Value, error) {
	var (
		gotCorrectAns = 0
		done          = make(chan struct{})
		ch            = make(chan qReadRes)
		values        = make(map[string]kv.Value, q.cfg.QuorumCfg.Read)
		best          kv.Value
	)

	for addr, rep := range q.replicas {
		rep := rep
		addr := addr
		go func() {
			rctx, cancel := context.WithTimeout(context.TODO(), q.cfg.OpTimeout)
			defer cancel()

			val, err := rep.Read(rctx, key)
			select {
			case ch <- qReadRes{
				err:  err,
				addr: addr,
				val:  val,
			}:
			case <-done:
				if err != nil {
					if !kv.IsNotFound(err) {
						q.log.Warnf(ctx, "not awaited: failed to read from node %s: %v", addr, err)
					}
					return
				}
				q.log.Debugf(ctx, "not awaited: successful read from node %s", addr)
			case <-ctx.Done():
			}
		}()
	}

await_loop:
	for i := 0; i < len(q.replicas); i++ {
		select {
		case readRes := <-ch:
			if readRes.err != nil && !kv.IsNotFound(readRes.err) {
				q.log.Warnf(ctx, "failed to read from node %s: %v", readRes.addr, readRes.err)
				continue await_loop
			}
			if !best.Ver.IsBigger(readRes.val.Ver) {
				best = readRes.val
			}
			gotCorrectAns++
			values[readRes.addr] = readRes.val
			if gotCorrectAns == q.cfg.QuorumCfg.Read {
				close(done)
				break await_loop
			}
		case <-ctx.Done():
			return nil, kv.Value{}, fmt.Errorf("%w: %w", replicator.ErrReadPhase, ctx.Err())
		}
	}
	if len(values) < q.cfg.QuorumCfg.Read {
		return nil, kv.Value{}, fmt.Errorf(
			"%w: %w: %d < %d",
			replicator.ErrReadPhase, replicator.ErrQuorumNotGathered,
			len(values),
			q.cfg.QuorumCfg.Read,
		)
	}
	return values, best, nil
}
