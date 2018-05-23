package seq

type rangeSeq struct {
	start, end, step int
}

func rangeNew(start, end, step int) Sequence {
	switch {
	case step > 0:
		if start >= end {
			return nil
		}
	case step < 0:
		if start <= end {
			return nil
		}
	default: //step == 0
		if start == end {
			return nil
		}
	}
	return &rangeSeq{
		start: start,
		end:   end,
		step:  step,
	}
}

func (s *rangeSeq) First() interface{} {
	return s.start
}

func (s *rangeSeq) Next() Sequence {
	new := rangeNew(s.start+s.step, s.end, s.step)
	if new == nil {
		return nil
	}
	return new
}

func (s *rangeSeq) String() string {
	return seqString(s)
}
