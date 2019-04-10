// Package seq is a lazy sequence library inspired by Clojure's
// sequence library. It uses transducers behind the scenes to
// implement the various functions. This provides a uniform mechanism
// for seperating the concept of a transformation from the mechanisms
// that perform that transformation.
package seq

import (
	"fmt"
	"reflect"
	"strings"

	"jsouthworth.net/go/dyn"
	"jsouthworth.net/go/transduce"
)

// Sequence is any type that can return iterate down its elements.
type Sequence interface {
	First() interface{}
	Next() Sequence
}

// Seqable is any type that can return a sequence to iterate over its elements.
type Seqable interface {
	Seq() Sequence
}

// First returns the first element of a sequence.
// coll is any type that can be converted to a Sequence by Seq.
func First(coll interface{}) interface{} {
	s := Seq(coll)
	if s == nil {
		return nil
	}
	return s.First()
}

// Next returns the sequence without the first element.
// coll is any type that can be converted to a Sequence by Seq.
func Next(coll interface{}) Sequence {
	s := Seq(coll)
	if s == nil {
		return nil
	}
	return s.Next()
}

// Conj conjoins a new element into a collection returning the
// new collection.
func Conj(coll interface{}, elem interface{}) interface{} {
	type conjoiner interface {
		Conj(elem interface{}) interface{}
	}
	switch v := coll.(type) {
	case conjoiner:
		return v.Conj(elem)
	default:
		val := reflect.ValueOf(coll)
		switch val.Kind() {
		case reflect.Slice:
			return sliceConj(val, elem)
		case reflect.Map:
			return mapConj(val, elem)
		default:
			_ = coll.(conjoiner)
			return nil
		}
	}
}

func sliceConj(coll reflect.Value, elem interface{}) interface{} {
	return reflect.Append(coll, reflect.ValueOf(elem)).Interface()
}

func mapConj(coll reflect.Value, elem interface{}) interface{} {
	entry := elem.(interface {
		Key() interface{}
		Value() interface{}
	})
	coll.SetMapIndex(reflect.ValueOf(entry.Key()),
		reflect.ValueOf(entry.Value()))
	return coll.Interface()
}

// Into takes an initial collection and a sequence and puts all
// the elements of the sequence into the collection returning the
// result.
func Into(to interface{}, from interface{}) interface{} {
	return Reduce(Conj, to, from)
}

// TransformInto takes an initial collection and a sequence and runs all
// the elements of the sequence through the transducer and places the
// results into the collection returning the result.
func TransformInto(
	to interface{},
	xfrm transduce.Transducer,
	from interface{},
) interface{} {
	return Transduce(xfrm, Conj, to, from)
}

// Transduce is a version of Reduce that takes a transducer and a
// reducing function and combines such that the transform is called on
// the elements being reduced. This is then passed to reduce to perform
// the actions greedily. The reducing function 'rf' must take both the
// result and the input type. That is, be of the type
// func(result rT, input eT) rT. This will be called using reflection
// unless it is the non-specialized
// func(result, input interface{})interface{}.
// coll is any type that can be converted to a Sequence by Seq.
func Transduce(
	xf transduce.Transducer,
	rf interface{},
	init interface{},
	coll interface{},
) interface{} {
	var rfunc func(result, input interface{}) interface{}
	switch f := rf.(type) {
	case func(result, input interface{}) interface{}:
		rfunc = f
	default:
		rfunc = func(result, input interface{}) interface{} {
			return apply(f, result, input)
		}
	}
	f := xf(transduce.Completing(rfunc))
	ret := Reduce(f.Step, init, coll)
	return f.Result(ret)
}

// Map returns a lazy sqeuence that contains the result of applying fn
// to each item in the Sequence. The transforming function 'fn' must match
// the signature func(in iT) oT and will be called using reflection unless
// it us the non-specialized type func(interface{})interface{}. coll is any
// type that can be converted to a Sequence by Seq.
func Map(fn interface{}, coll interface{}) Sequence {
	return XfrmSequence(transduce.Map(fn), Seq(coll))
}

// Replace returns a lazy sequence that contains the result of replacing
// the values in the provided smap for the ones in the sequence. smap must
// be one of the following types something that implements
// interface { Find(interface{}) (interface{},bool) }, map[iT]oT. Reflection
// is used unless the map is of the non specialized map[interface{}]interface{}
// type. coll is any type that can be converted to a Sequence by Seq.
func Replace(smap interface{}, coll interface{}) Sequence {
	return XfrmSequence(transduce.Replace(smap), Seq(coll))
}

// Reduce takes a function and iterates over the sequence calling the
// function with the element at that place in the sequence and the result
// of the previous call. The initial result is provided as 'init' to the
// Reduce function. The reducing function 'fn' must match the signature
// func(result rT, input iT) rT and will be called using reflection unless
// is is the non-specialized type func(result, input interface{})interface{}.
// coll is any type that can be converted to a Sequence by Seq.
func Reduce(
	fn interface{},
	init interface{},
	coll interface{},
) interface{} {
	f := wrapReduce(fn)
	//TODO: make a reducer interface to make this efficient
	s := Seq(coll)
	if s == nil {
		return init
	}
	var ret interface{} = init
	for s != nil {
		ret = f(ret, First(s))
		s = Seq(Next(s))
	}
	return ret
}

func wrapReduce(f interface{}) func(res, in interface{}) interface{} {
	switch fn := f.(type) {
	case func(interface{}, interface{}) interface{}:
		return fn
	default:
		return func(result, input interface{}) interface{} {
			return apply(f, result, input)
		}
	}
}

// RangeUntil returns a lazy sequence that will be the integers [0,end).
func RangeUntil(end int) Sequence {
	return Range(0, end, 1)
}

// RangeBetween returns a lazy sequence that will be the integers [start,end).
func RangeBetween(start, end int) Sequence {
	return Range(start, end, 1)
}

// Range returns a lazy sequence that will be the integers
// [start, start+step, ..., end)
func Range(start, end, step int) Sequence {
	return rangeNew(start, end, step)
}

// Repeat will return a lazy sequence that repeats x, n times.
func Repeat(n int, x interface{}) Sequence {
	return repeatSeqNew(n, x)
}

// RepeatInfinitely will return a lazy sequence that repeats x forever.
func RepeateInfinitely(x interface{}) Sequence {
	return infiniteRepeatSeq(x)
}

// Iterate will return the result of calling fn on the result of the previous
// call of fn. The iteration starts with the passed in x.
func Iterate(fn interface{}, x interface{}) Sequence {
	return iterateNew(fn, x)
}

// Take will return a lazy but finite sequence consisting of the first
// n elements of the passed in sequence. coll is any type that can be
// converted to a Sequence by Seq.
func Take(n int, coll interface{}) Sequence {
	return XfrmSequence(transduce.Take(n), Seq(coll))
}

// TakeNth will return a lazy sequence consisting of every nth item of
// the passed in collection. coll is any type that can be converted to
// a Sequence by Seq.
func TakeNth(n int, coll interface{}) Sequence {
	return XfrmSequence(transduce.TakeNth(n), Seq(coll))
}

// Drop returns a lazy sequence that contains all but the first
// n elements in the passed in sequence. coll is any type that can
// be converted to a Sequence by Seq.
func Drop(n int, coll interface{}) Sequence {
	return XfrmSequence(transduce.Drop(n), Seq(coll))
}

// Cycle returns a lazy sequence consisting of the repeating the
// elements of coll. coll is any type that can be converted to a
// Sequence by Seq.
func Cycle(coll interface{}) Sequence {
	return cycleSeq(Seq(coll))
}

// Interleave returns a lazy sequence of the first element of each
// passed in sequence followed by the second, followed by the third, and so on.
// coll is any type that can be converted to a Sequence by Seq.
//
//  [coll[0][0], coll[1][0], ..., coll[n][0], ...,
//   coll[0][m], coll[1[m], ..., coll[n][m]]
func Interleave(colls ...interface{}) Sequence {
	return LazySeq(func() Sequence {
		for i, coll := range colls {
			colls[i] = Seq(coll)
			if colls[i] == nil {
				return nil
			}
		}
		rests := make([]interface{}, len(colls))
		var nomore bool
		for i, coll := range colls {
			rests[i] = Next(coll)
			if rests[i] == nil {
				nomore = true
			}
		}
		out := Interleave(rests...)
		if nomore {
			out = nil
		}
		for i := len(colls) - 1; i >= 0; i-- {
			out = Cons(First(colls[i]), out)
		}
		return out
	})
}

// Interpose returns a lazy sequence of  the elements of the passed in sequence
// seperated by the passed in seperator. coll is any type that can be converted
// to a Sequence by Seq.
func Interpose(seperator interface{}, coll interface{}) Sequence {
	return XfrmSequence(transduce.Interpose(seperator), Seq(coll))
}

// Filter returns a lazy sequence that will contain the elements of the
// passed in sequence for which pred is true. pred must match the signature
// func(i iT) bool and will be called with reflection unless it is the
// non-specialized type func(interface{}) bool. coll is any type that can
// be converted to a Sequence by Seq.
func Filter(pred interface{}, coll interface{}) Sequence {
	return XfrmSequence(transduce.Filter(pred), Seq(coll))
}

// Filter returns a lazy sequence that will contain the elements of the
// passed in sequence for which pred is false. pred must match the signature
// func(i iT) bool and will be called with reflection unless it is the
// non-specialized type func(interface{}) bool. coll is any type that can
// be converted to a Sequence by Seq.
func Remove(pred interface{}, coll interface{}) Sequence {
	return XfrmSequence(transduce.Remove(pred), Seq(coll))
}

// TakeWhile returns a lazy sequence of the items from the passed in sequence
// so long as pred returns true. pred must match the signature
// func(i iT) bool and will be called with reflection unless it is the
// non-specialized type func(interface{}) bool. coll is any type that can
// be converted to a Sequence by Seq.
func TakeWhile(pred interface{}, coll interface{}) Sequence {
	return XfrmSequence(transduce.TakeWhile(pred), Seq(coll))
}

// DropWhile returns a lazy sequence of the items from the passed in sequence
// starting with the first element that for which pred returns false.
// pred must match the signature func(i iT) bool and will be called with
// reflection unless it is the non-specialized type func(interface{}) bool.
// coll is any type that can be converted to a Sequence by Seq.
func DropWhile(pred interface{}, coll interface{}) Sequence {
	return XfrmSequence(transduce.DropWhile(pred), Seq(coll))
}

// Keep returns a lazy sequence for which f returns a non nil value
// The function f must be of the type func(i iT) oT and will be
// called with reflection unless it is the non-specialized type
// func(interface{}) interface{}. coll is any type that can be
// converted to a Sequence by Seq.
func Keep(f interface{}, coll interface{}) Sequence {
	return XfrmSequence(transduce.Keep(f), Seq(coll))
}

// KeepIndexed returns a lazy sequence for which f returns a non nil value
// The function f must be of the type func(idx int, i iT) oT and will be
// called with reflection unless it is the non-specialized type
// func(int, interface{}) interface{}. coll is any type that can be
// converted to a Sequence by Seq.
func KeepIndexed(f interface{}, coll interface{}) Sequence {
	return XfrmSequence(transduce.KeepIndexed(f), Seq(coll))
}

// Dedupe returns a lazy sequence with any duplicates removed.
// coll is any type that can be converted to a Sequence by Seq.
func Dedupe(coll interface{}) Sequence {
	return XfrmSequence(transduce.Dedupe(), Seq(coll))
}

// SplitAt returns a sequence containing two sequences corresponding
// to the split index. coll is any type that can be converted to a
// Sequence by Seq.
func SplitAt(index int, coll interface{}) Sequence {
	s := Seq(coll)
	return Cons(Take(index, s),
		Cons(Drop(index, s), nil))
}

// SplitWith returns a sequence containing two sequences corresponding
// to predicate.
// pred must match the signature func(i iT) bool and will be called with
// reflection unless it is the non-specialized type func(interface{}) bool.
// coll is any type that can be converted to a Sequence by Seq.
func SplitWith(pred interface{}, coll interface{}) Sequence {
	s := Seq(coll)
	return Cons(TakeWhile(pred, s),
		Cons(DropWhile(pred, s), nil))
}

// Every will iterate over every element of the sequence and return if
// the predicate hold for every element. pred must match the signature
// func(i iT) bool and will be called with reflection unless it is the
// non-specialized type func(interface{}) bool. coll is any type that
// can be converted to a Sequence by Seq.
func Every(pred interface{}, coll interface{}) bool {
	s := Seq(coll)
	for {
		switch {
		case s == nil:
			return true
		case apply(pred, First(s)).(bool):
			s = Next(s)
		default:
			return false
		}
	}
}

// Some will return if pred is true for some element in the sequence.
// pred must match the signature func(i iT) bool and will be called with
// reflection unless it is the non-specialized type func(interface{}) bool.
// coll is any type that can be converted to a Sequence by Seq.
func Some(pred interface{}, coll interface{}) bool {
	s := Seq(coll)
	if s == nil {
		return false
	}
	return apply(pred, First(s)).(bool) || Some(pred, Next(s))
}

// NotEvery is the inverse of Every.
// pred must match the signature func(i iT) bool and will be called with
// reflection unless it is the non-specialized type func(interface{}) bool.
// coll is any type that can be converted to a Sequence by Seq.
func NotEvery(pred interface{}, coll interface{}) bool {
	return !Every(pred, coll)
}

// NotAny is the inverse of Some.
// pred must match the signature func(i iT) bool and will be called with
// reflection unless it is the non-specialized type func(interface{}) bool.
// coll is any type that can be converted to a Sequence by Seq.
func NotAny(pred interface{}, coll interface{}) bool {
	return !Some(pred, coll)
}

// DoAll will realize every element in a lazy sequence and return that sequence.
// coll is any type that can be converted to a Sequence by Seq.
func DoAll(coll interface{}) Sequence {
	s := Seq(coll)
	DoRun(s)
	return s
}

// DoRun will realize every element in a lazy sequence.
// coll is any type that can be converted to a Sequence by Seq.
func DoRun(coll interface{}) {
	s := Seq(coll)
	for s != nil {
		s = Seq(Next(s))
	}
}

// Seq will convert a type to a sequence. If the type is Sequable it will
// run Seq(), if it is already a sequence it will return the sequence,
// otherwise it will attempt to build a sequence using reflection.
// Currently it supports automatic conversion of both
// arbitray go slices ([]T) and strings.
func Seq(coll interface{}) Sequence {
	if coll == nil {
		return nil
	}
	switch seq := coll.(type) {
	case Seqable:
		return seq.Seq()
	case Sequence:
		return seq
	default:
		return reflectSeq(coll)
	}
}

// Slice will convert a lazy sequence to a go slice realizing each element.
// coll is any type that can be converted to a Sequence by Seq.
func Slice(coll interface{}) []interface{} {
	return Reduce(func(a []interface{}, b interface{}) []interface{} {
		return append(a, b)
	}, []interface{}{}, coll).([]interface{})
}

// Concat returns a lazy sequence that is the concatenation of the provided
// sequences. coll is any type that can be converted to a Sequence by Seq.
func Concat(colls ...interface{}) Sequence {
	return XfrmSequence(transduce.Cat(Reduce), Seq(colls))
}

// Mapcat returns a lazy sequence that is the concatenation of the
// provided sequences modified by the mapping function f.
// f must be of the form func(in iT) oT and will be called with
// reflection unless it is the non-specialized func(interface{})interface{}.
// colls is an type that can be converted to a Sequence by Seq.
func Mapcat(f interface{}, colls ...interface{}) Sequence {
	return XfrmSequence(transduce.Mapcat(Reduce, f), Seq(colls))
}

// PartitionBy returns a lazy sequence that consists of partitions of
// the provided sequence. The partitions are determined by f which is
// any function of type func(i iT) oT. When f returns a different value
// from its previous call then a parition is created.
// coll is any type that can be converted to a Sequence by Seq.
func PartitionBy(f interface{}, coll interface{}) Sequence {
	return XfrmSequence(
		transduce.Compose(transduce.PartitionBy(f),
			transduce.Map(Seq)),
		Seq(coll))
}

// PartitionAll returns a lazy sequence that consists of partitions of
// the provided sequence of n elements. If the length of the sequence
// is not a multiple of n then the remainder will be returned as the
// last element. coll is any type that can be converted to a Sequence by Seq.
func PartitionAll(n int, coll interface{}) Sequence {
	return XfrmSequence(
		transduce.Compose(transduce.PartitionAll(n),
			transduce.Map(Seq)),
		Seq(coll))
}

// apply allows one to call arbitrary go functions using reflection.
// It handles the pitfalls of calling functions using reflection
// so that a simple interface is provided to callers.
func apply(f interface{}, args ...interface{}) interface{} {
	return dyn.Apply(f, args...)
}

// ConvertToString converts any Sequence to a string. This is useful for
// other sequence implementations that would like to use the same
// algorithm.
func ConvertToString(coll Sequence) string {
	return seqString(coll)
}

func seqString(coll Sequence) string {
	var b strings.Builder
	coll = Seq(coll)
	if coll == nil {
		return "()"
	}
	fmt.Fprint(&b, "(")
	for coll != nil {
		first := First(coll)
		next := Seq(Next(coll))
		if next == nil {
			fmt.Fprintf(&b, "%v", first)
		} else {
			fmt.Fprintf(&b, "%v ", first)
		}
		coll = next
	}
	fmt.Fprint(&b, ")")
	return b.String()
}
