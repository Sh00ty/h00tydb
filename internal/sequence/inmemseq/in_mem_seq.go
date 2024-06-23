package inmemseq

import "sync/atomic"

type Sequence struct {
	seq atomic.Uint64
}

func NewInMemSeq() *Sequence {
	return &Sequence{}
}

func (s *Sequence) Inc() uint64 {
	return s.seq.Add(1)
}
func (s *Sequence) Get() uint64 {
	return s.seq.Load()
}
