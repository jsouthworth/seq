package seq

type cons struct {
	first interface{}
	next  Sequence
}

func (s *cons) First() interface{} {
	return s.first
}
func (s *cons) Next() Sequence {
	return s.next
}

func (s *cons) String() string {
	return seqString(s)
}

func consNew(first interface{}, next Sequence) *cons {
	return &cons{first: first, next: next}
}

func Cons(v interface{}, coll Sequence) Sequence {
	return consNew(v, coll)
}

func Empty() Sequence {
	return nil
}
