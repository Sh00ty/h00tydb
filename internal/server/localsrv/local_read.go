package localsrv

import (
	"context"

	"gitlab.com/Sh00ty/hootydb/internal/kv"
	"gitlab.com/Sh00ty/hootydb/internal/transport/transport"
	"gitlab.com/Sh00ty/hootydb/internal/utils/random"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Srv) LocalRead(ctx context.Context, in *transport.LocalReadRequest) (out *transport.LocalReadResponse, err error) {
	if random.Random(100) <= s.faultPercent {
		return &transport.LocalReadResponse{}, status.Error(codes.Internal, "rpc fault injection")
	}

	val, err := s.storage.Get(ctx, kv.Key(in.Key))
	if err != nil {
		if kv.IsNotFound(err) {
			return &transport.LocalReadResponse{}, status.Errorf(codes.NotFound, "not found key %s", in.Key)
		}
		return &transport.LocalReadResponse{}, status.Errorf(codes.Internal, "internal kv store error %v on key %s", err, in.Key)
	}

	return &transport.LocalReadResponse{
		Value: &transport.Value{
			Val: val.Val.(string),
			Ver: val.Ver.String(),
		},
	}, nil
}
