package seq

import (
	"sync"
)

type iterate struct {
	mu        sync.Mutex
	realized  bool
	cur, prev interface{}
	fn        interface{}
	next      *iterate
}

func iterateNew(fn interface{}, x interface{}) *iterate {
	return &iterate{
		fn:       fn,
		cur:      x,
		realized: true,
	}
}

func (s *iterate) First() interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.first()
}

func (s *iterate) first() interface{} {
	if !s.realized {
		s.cur = apply(s.fn, s.prev)
	}
	return s.cur
}

func (s *iterate) Next() Sequence {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.next == nil {
		s.next = &iterate{
			fn:   s.fn,
			prev: s.first(),
		}
	}
	return s.next
}

func (s *iterate) String() string {
	return seqString(s)
}
