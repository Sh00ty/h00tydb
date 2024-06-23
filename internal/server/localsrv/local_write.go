package localsrv

import (
	"context"
	"errors"

	"gitlab.com/Sh00ty/hootydb/internal/kv"
	"gitlab.com/Sh00ty/hootydb/internal/transport/transport"
	"gitlab.com/Sh00ty/hootydb/internal/utils/random"
	"gitlab.com/Sh00ty/hootydb/internal/versions"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Srv) LocalWrite(ctx context.Context, in *transport.LocalWriteRequest) (out *transport.LocalWriteResponse, err error) {
	if random.Random(100) <= s.faultPercent {
		return &transport.LocalWriteResponse{}, status.Error(codes.Internal, "rpc fault injection")
	}
	ver, err := s.verBuilder.FromString(in.Value.Ver)
	if err != nil {
		return &transport.LocalWriteResponse{}, status.Errorf(codes.InvalidArgument, "invalid version format: %v", err)
	}
	err = s.storage.Set(ctx, kv.Key(in.Key), kv.Value{
		Val: in.Value.Val,
		Ver: ver,
	})
	if err != nil {
		e := &kv.KVError{}
		if errors.As(err, &e) && e.Kind == kv.OldVer {
			newVer := e.Extra[1]
			return &transport.LocalWriteResponse{
				IsOldVer: true,
				NewVer:   newVer.(versions.Version).String(),
			}, nil
		}
		return &transport.LocalWriteResponse{}, status.Errorf(codes.Internal, "internal error from storage on key %s: %v", in.Key, err)
	}

	return &transport.LocalWriteResponse{
		Ok: true,
	}, nil
}
