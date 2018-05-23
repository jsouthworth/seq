package seq

type cycle struct {
	all Sequence
	seq Sequence
}

func cycleSeq(all Sequence) *cycle {
	return &cycle{all: all, seq: all}
}

func (c *cycle) First() interface{} {
	return c.seq.First()
}

func (c *cycle) Next() Sequence {
	nxt := Seq(Next(c.seq))
	if nxt == nil {
		nxt = c.all
	}
	return &cycle{all: c.all, seq: nxt}
}

func (c *cycle) String() string {
	return seqString(c)
}
