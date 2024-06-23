package quorum

import (
	"context"
	"fmt"

	"gitlab.com/Sh00ty/hootydb/internal"
	"gitlab.com/Sh00ty/hootydb/internal/config"
	"gitlab.com/Sh00ty/hootydb/internal/kv"
	"gitlab.com/Sh00ty/hootydb/internal/replicator"
)

type QWriter struct {
	cfg      config.Config
	log      internal.Logger
	replicas map[string]replicator.Client
}

type qWriteRes struct {
	addr string
	err  error
}

type WriteRes struct {
	Addr     string
	IsOldVer bool
}

// все реплики, которых нет в replicas уже успешно ответили
func (q *QWriter) WritePhase(ctx context.Context, key kv.Key, val kv.Value, replicas []string) (map[string]WriteRes, error) {
	var (
		gotCorrectAns = len(q.replicas) - len(replicas)
		done          = make(chan struct{})
		ch            = make(chan qWriteRes)
		wrote         = make(map[string]WriteRes, q.cfg.QuorumCfg.Write)
	)

	for _, addr := range replicas {
		addr := addr
		rep := q.replicas[addr]
		go func() {
			wctx, cancel := context.WithTimeout(context.TODO(), q.cfg.OpTimeout)
			defer cancel()

			err := rep.Write(wctx, key, val)
			select {
			case ch <- qWriteRes{
				addr: addr,
				err:  err,
			}:
			case <-done:
				if err != nil {
					if !kv.IsOldVer(err) {
						q.log.Warnf(ctx, "not awaited: failed to write on node %s: %v", addr, err)
					}
					return
				}
				q.log.Debugf(ctx, "not awaited: successful write on node %s", addr)
			case <-ctx.Done():
			}
		}()
	}

await_loop:
	for i := 0; i < len(replicas); i++ {
		select {
		case writeRes := <-ch:
			var isOldVer bool
			if writeRes.err != nil {
				if !kv.IsOldVer(writeRes.err) {
					q.log.Warnf(ctx, "failed to write on node %s: %v", writeRes.addr, writeRes.err)
					continue
				}
				isOldVer = true
			}
			gotCorrectAns++
			wrote[writeRes.addr] = WriteRes{
				Addr:     writeRes.addr,
				IsOldVer: isOldVer,
			}
			if gotCorrectAns == q.cfg.QuorumCfg.Write {
				close(done)
				break await_loop
			}
		case <-ctx.Done():
			return nil, fmt.Errorf("%w: %w", replicator.ErrWritePhase, ctx.Err())
		}
	}
	if gotCorrectAns < q.cfg.QuorumCfg.Write {
		return nil, fmt.Errorf(
			"%w: %w: %d < %d",
			replicator.ErrWritePhase,
			replicator.ErrQuorumNotGathered,
			len(wrote),
			q.cfg.QuorumCfg.Write,
		)
	}
	return wrote, nil
}
