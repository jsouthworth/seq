package seq

import (
	"sync"

	"jsouthworth.net/go/transduce"
)

type xfrmSeq struct {
	mu sync.Mutex

	buffer       *buffer
	step         transduce.ReducerFn
	coll         Sequence
	bufferedColl Sequence
	completed    bool
}

// XfrmSequence returns a lazy sequence that is the result of stepping
// the transducer over the elements of the passed in sequence. The
// resulting sequence may be longer or shorter than the original sequence
// based on the results of the transducer.
func XfrmSequence(xf transduce.Transducer, coll Sequence) Sequence {
	if coll == nil {
		return nil
	}
	buffer := &buffer{}
	ret := &xfrmSeq{
		coll:   coll,
		buffer: buffer,
	}
	ret.step = xf(transduce.Completing(
		func(result, input interface{}) interface{} {
			if !transduce.IsReduced(result) {
				buffer.add(input)
			}
			return result
		}))
	return ret
}

func (s *xfrmSeq) Seq() Sequence {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.bufferedColl != nil || s.completed {
		return s
	}
	/*
	   This process should cache the state for this element in the
	   sequence and then always return that cached state

	   The cached state should be the first value, the next
	   Sequence, whether the sequence completed successfully.

	   When the process ends, Complete must be called exactly
	   once.  That is, if coll == nil, call Complete once, or if
	   result is reduced call complete.

	   The next sequence should be the buffered values with the
	   tail set to the next of the original collection.

	*/
	coll := s.coll
	for s.bufferedColl == nil {
		res := s.step.Step(nil, First(coll))
		coll = Next(coll)
		if s.buffer.head != nil {
			if coll != nil && !transduce.IsReduced(res) {
				s.buffer.tail.next = &xfrmSeq{
					step:   s.step,
					coll:   coll,
					buffer: s.buffer,
				}
			}
			s.bufferedColl = s.buffer.head
			s.buffer.clear()
			s.buffer = nil
		}
		if transduce.IsReduced(res) {
			s.step.Result(nil)
			s.completed = true
			break
		}
		if coll == nil {
			s.step.Result(nil)
			if s.buffer != nil && s.buffer.head != nil {
				s.bufferedColl = s.buffer.head
				s.buffer.clear()
				s.buffer = nil
			}
			s.completed = true
			break
		}

	}
	if s.completed && s.bufferedColl == nil {
		return nil
	}
	return s
}

func (s *xfrmSeq) First() interface{} {
	s.Seq()
	return First(s.bufferedColl)
}

func (s *xfrmSeq) Next() Sequence {
	s.Seq()
	return Next(s.bufferedColl)
}

func (s *xfrmSeq) String() string {
	return seqString(s)
}

type buffer struct {
	head   *cons
	tail   *cons
	length int
}

func (b *buffer) clear() *buffer {
	b.head = nil
	b.tail = nil
	b.length = 0
	return b
}

func (b *buffer) add(item interface{}) *buffer {
	l := &cons{first: item}
	if b.length == 0 {
		b.head, b.tail = l, l
	} else {
		b.tail.next = &cons{first: item}
		b.tail = b.tail.next.(*cons)
	}
	b.length++
	return b
}
