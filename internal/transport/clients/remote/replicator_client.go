package remote

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/Sh00ty/hootydb/internal/kv"
	"gitlab.com/Sh00ty/hootydb/internal/transport/transport"
	"gitlab.com/Sh00ty/hootydb/internal/versions"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// TODO: error handling with retries

var callOpts = []grpc.CallOption{
	grpc.WaitForReady(true),
}

type ReplicatorClient struct {
	clnt       transport.ReplicatorServiceClient
	verBuilder versions.VersionBuilder
	addr       string
}

func NewReplicatorClient(addr string, verBuilder versions.VersionBuilder) (*ReplicatorClient, error) {
	cc, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff:           backoff.DefaultConfig,
			MinConnectTimeout: time.Second,
		}),
	)
	if err != nil {
		return nil, err
	}

	return &ReplicatorClient{
		addr:       addr,
		verBuilder: verBuilder,
		clnt:       transport.NewReplicatorServiceClient(cc),
	}, nil
}

func (c *ReplicatorClient) Read(ctx context.Context, key kv.Key) (val kv.Value, err error) {

	out, err := c.clnt.LocalRead(ctx, &transport.LocalReadRequest{
		Key: string(key),
	}, callOpts...)

	if err != nil {
		switch status.Code(err) {
		case codes.NotFound:
			return kv.Value{}, kv.MakeNotFound(key)
		}
		return kv.Value{}, err
	}
	val.Val = out.Value.Val

	ver, err := c.verBuilder.FromString(out.Value.Ver)
	if err != nil {
		return kv.Value{}, err
	}
	val.Ver = ver
	return
}

func (c *ReplicatorClient) Write(ctx context.Context, key kv.Key, val kv.Value) error {
	in := &transport.LocalWriteRequest{
		Key: string(key),
		Value: &transport.Value{
			Val: val.Val.(string),
			Ver: val.Ver.String(),
		},
	}
	out, err := c.clnt.LocalWrite(ctx, in, callOpts...)
	if err != nil {
		return err
	}
	if out.IsOldVer {
		newVer, _ := c.verBuilder.FromString(out.NewVer)
		return kv.MakeOldVerError(key, val.Ver, newVer)
	}
	if !out.Ok {
		return fmt.Errorf("got unknown error %w on local write to %s", err, c.addr)
	}
	return nil
}
