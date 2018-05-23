package seq

import (
	"fmt"
	"reflect"
)

type sliceSeq struct {
	v reflect.Value
}

func (s sliceSeq) First() interface{} {
	return s.v.Index(0).Interface()
}

func (s sliceSeq) Next() Sequence {
	if s.v.Len() <= 1 {
		return nil
	}
	return sliceSeq{v: s.v.Slice(1, s.v.Len())}
}

func (s sliceSeq) String() string {
	return seqString(s)
}

func reflectSeq(coll interface{}) Sequence {
	v := reflect.ValueOf(coll)
	switch v.Kind() {
	case reflect.Slice:
		return sliceSequence(v)
	case reflect.String:
		return sliceSequence(reflect.ValueOf([]rune(coll.(string))))
	default:
		panic(fmt.Errorf("cannot convert %T to Seq", coll))
	}
}

func sliceSequence(v reflect.Value) Sequence {
	if v.Len() == 0 {
		return nil
	}
	return sliceSeq{v: v}
}
