package fixed

import "context"

type FixedResolver struct {
	self   string
	others []string
}

func NewFromSlice(self string, addrs ...string) *FixedResolver {
	return &FixedResolver{self: self, others: addrs}
}

func (r *FixedResolver) GetRemoteAddrs(ctx context.Context) []string {
	return r.others
}

func (r *FixedResolver) GetSelf(ctx context.Context) string {
	return r.self
}
