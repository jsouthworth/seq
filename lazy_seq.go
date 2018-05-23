package seq

import (
	"sync"
)

type lazySeq struct {
	mu  sync.Mutex
	fn  func() Sequence
	seq Sequence
}

func LazySeq(fn func() Sequence) Sequence {
	return &lazySeq{fn: fn}
}

func (s *lazySeq) Seq() Sequence {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.fn != nil {
		s.seq = s.fn()
		s.fn = nil
	}

	var ls Sequence = s.seq
	tmp, ok := ls.(*lazySeq)
	for ok {
		ls = tmp.Seq()
		tmp, ok = ls.(*lazySeq)
	}
	s.seq = ls

	return s.seq
}
func (s *lazySeq) First() interface{} {
	s.Seq()
	if s.seq == nil {
		return nil
	}
	return First(s.seq)
}
func (s *lazySeq) Next() Sequence {
	s.Seq()
	if s.seq == nil {
		return nil
	}
	return Next(s.seq)
}

func (s *lazySeq) String() string {
	return seqString(s)
}
