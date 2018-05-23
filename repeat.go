package seq

const inf = -1

type repeatSeq struct {
	count int
	val   interface{}
}

func repeatSeqNew(count int, val interface{}) Sequence {
	return &repeatSeq{
		count: count,
		val:   val,
	}
}

func infiniteRepeatSeq(val interface{}) Sequence {
	return repeatSeqNew(inf, val)
}

func (s *repeatSeq) First() interface{} {
	return s.val
}

func (s *repeatSeq) Next() Sequence {
	switch {
	case s.count > 1:
		return repeatSeqNew(s.count-1, s.val)
	case s.count == inf:
		return s
	default:
		return nil
	}
}

func (s *repeatSeq) String() string {
	return seqString(s)
}
