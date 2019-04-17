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
	case reflect.Map:
		return mapSequence(v)
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

// MapEntry is a key,value pair representing an item in a map
// when treated as a sequence.
type MapEntry interface {
	Key() interface{}
	Value() interface{}
}

type mapEntry struct {
	key interface{}
	val interface{}
}

func (e mapEntry) Key() interface{} {
	return e.key
}

func (e mapEntry) Value() interface{} {
	return e.val
}

type mapSeq struct {
	keys []reflect.Value
	m    reflect.Value
}

func (s mapSeq) First() interface{} {
	k := s.keys[0]
	v := s.m.MapIndex(k)
	return mapEntry{
		key: k.Interface(),
		val: v.Interface(),
	}
}

func (s mapSeq) Next() Sequence {
	if len(s.keys) == 1 {
		return nil
	}
	return mapSeq{
		keys: s.keys[1:],
		m:    s.m,
	}
}

func mapSequence(v reflect.Value) Sequence {
	if v.Len() == 0 {
		return nil
	}
	return mapSeq{
		keys: v.MapKeys(),
		m:    v,
	}
}
