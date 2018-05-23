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
func First(coll Sequence) interface{} {
	if coll == nil {
		return nil
	}
	return coll.First()
}

// Next returns the sequence without the first element.
func Next(coll Sequence) Sequence {
	if coll == nil {
		return nil
	}
	return coll.Next()
}

// Transduce is a version of Reduce that takes a transducer and a
// reducing function and combines such that the transform is called on
// the elements being reduced. This is then passed to reduce to perform
// the actions greedily. The reducing function 'rf' must take both the
// result and the input type. That is, be of the type
// func(result rT, input eT) rT. This will be called using reflection
// unless it is the non-specialized
// func(result, input interface{})interface{}.
func Transduce(
	xf transduce.Transducer,
	rf interface{},
	init interface{},
	coll Sequence,
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
// it us the non-specialized type func(interface{})interface{}.
func Map(fn interface{}, coll Sequence) Sequence {
	return XfrmSequence(transduce.Map(fn), Seq(coll))
}

// Replace returns a lazy sequence that contains the result of replacing
// the values in the provided smap for the ones in the sequence. smap must
// be one of the following types something that implements
// interface { Find(interface{}) (interface{},bool) }, map[iT]oT. Reflection
// is used unless the map is of the non specialized map[interface{}]interface{}
// type.
func Replace(smap interface{}, coll Sequence) Sequence {
	return XfrmSequence(transduce.Replace(smap), Seq(coll))
}

// Reduce takes a function and iterates over the sequence calling the
// function with the element at that place in the sequence and the result
// of the previous call. The initial result is provided as 'init' to the
// Reduce function. The reducing function 'fn' must match the signature
// func(result rT, input iT) rT and will be called using reflection unless
// is is the non-specialized type func(result, input interface{})interface{}.
func Reduce(
	fn interface{},
	init interface{},
	coll Sequence,
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
// n elements of the passed in sequence.
func Take(n int, coll Sequence) Sequence {
	return XfrmSequence(transduce.Take(n), coll)
}

// TakeNth will return a lazy sequence consisting of every nth item of
// the passed in collection.
func TakeNth(n int, coll Sequence) Sequence {
	return XfrmSequence(transduce.TakeNth(n), coll)
}

// Drop returns a lazy sequence that contains all but the first
// n elements in the passed in sequence.
func Drop(n int, coll Sequence) Sequence {
	return XfrmSequence(transduce.Drop(n), coll)
}

// Cycle returns a lazy sequence consisting of the repeating the
// elements of coll.
func Cycle(coll Sequence) Sequence {
	return cycleSeq(coll)
}

// Interleave returns a lazy sequence of the first element of each
// passed in sequence followed by the second, followed by the third, and so on.
//
//  [coll[0][0], coll[1][0], ..., coll[n][0], ...,
//   coll[0][m], coll[1[m], ..., coll[n][m]]
func Interleave(colls ...Sequence) Sequence {
	return LazySeq(func() Sequence {
		for i, coll := range colls {
			colls[i] = Seq(coll)
			if colls[i] == nil {
				return nil
			}
		}
		rests := make([]Sequence, len(colls))
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
// seperated by the passed in seperator.
func Interpose(seperator interface{}, coll Sequence) Sequence {
	return XfrmSequence(transduce.Interpose(seperator), coll)
}

// Filter returns a lazy sequence that will contain the elements of the
// passed in sequence for which pred is true. pred must match the signature
// func(i iT) bool and will be called with reflection unless it is the
// non-specialized type func(interface{}) bool.
func Filter(pred interface{}, coll Sequence) Sequence {
	return XfrmSequence(transduce.Filter(pred), coll)
}

// Filter returns a lazy sequence that will contain the elements of the
// passed in sequence for which pred is false. pred must match the signature
// func(i iT) bool and will be called with reflection unless it is the
// non-specialized type func(interface{}) bool.
func Remove(pred interface{}, coll Sequence) Sequence {
	return XfrmSequence(transduce.Remove(pred), coll)
}

// TakeWhile returns a lazy sequence of the items from the passed in sequence
// so long as pred returns true. pred must match the signature
// func(i iT) bool and will be called with reflection unless it is the
// non-specialized type func(interface{}) bool.
func TakeWhile(pred interface{}, coll Sequence) Sequence {
	return XfrmSequence(transduce.TakeWhile(pred), coll)
}

// DropWhile returns a lazy sequence of the items from the passed in sequence
// starting with the first element that for which pred returns false.
// pred must match the signature func(i iT) bool and will be called with
// reflection unless it is the non-specialized type func(interface{}) bool.
func DropWhile(pred interface{}, coll Sequence) Sequence {
	return XfrmSequence(transduce.DropWhile(pred), coll)
}

// Keep returns a lazy sequence for which f returns a non nil value
// The function f must be of the type func(i iT) oT and will be
// called with reflection unless it is the non-specialized type
// func(interface{}) interface{}.
func Keep(f interface{}, coll Sequence) Sequence {
	return XfrmSequence(transduce.Keep(f), coll)
}

// KeepIndexed returns a lazy sequence for which f returns a non nil value
// The function f must be of the type func(idx int, i iT) oT and will be
// called with reflection unless it is the non-specialized type
// func(int, interface{}) interface{}.
func KeepIndexed(f interface{}, coll Sequence) Sequence {
	return XfrmSequence(transduce.KeepIndexed(f), coll)
}

// Dedupe returns a lazy sequence with any duplicates removed.
func Dedupe(coll Sequence) Sequence {
	return XfrmSequence(transduce.Dedupe(), coll)
}

// SplitAt returns a sequence containing two sequences corresponding
// to the split index.
func SplitAt(index int, coll Sequence) Sequence {
	return Cons(Take(index, coll),
		Cons(Drop(index, coll), nil))
}

// SplitWith returns a sequence containing two sequences corresponding
// to predicate.
// pred must match the signature func(i iT) bool and will be called with
// reflection unless it is the non-specialized type func(interface{}) bool.
func SplitWith(pred interface{}, coll Sequence) Sequence {
	return Cons(TakeWhile(pred, coll),
		Cons(DropWhile(pred, coll), nil))
}

// Every will iterate over every element of the sequence and return if
// the predicate hold for every element. pred must match the signature
// func(i iT) bool and will be called with reflection unless it is the
// non-specialized type func(interface{}) bool.
func Every(pred interface{}, coll Sequence) bool {
	for {
		switch {
		case Seq(coll) == nil:
			return true
		case apply(pred, First(coll)).(bool):
			coll = Next(coll)
		default:
			return false
		}
	}
}

// Some will return if pred is true for some element in the sequence.
// pred must match the signature func(i iT) bool and will be called with
// reflection unless it is the non-specialized type func(interface{}) bool.
func Some(pred interface{}, coll Sequence) bool {
	if Seq(coll) == nil {
		return false
	}
	return apply(pred, First(coll)).(bool) || Some(pred, Next(coll))
}

// NotEvery is the inverse of Every.
// pred must match the signature func(i iT) bool and will be called with
// reflection unless it is the non-specialized type func(interface{}) bool.
func NotEvery(pred interface{}, coll Sequence) bool {
	return !Every(pred, coll)
}

// NotAny is the inverse of Some.
// pred must match the signature func(i iT) bool and will be called with
// reflection unless it is the non-specialized type func(interface{}) bool.
func NotAny(pred interface{}, coll Sequence) bool {
	return !Some(pred, coll)
}

// DoAll will realize every element in a lazy sequence and return that sequence.
func DoAll(coll Sequence) Sequence {
	DoRun(coll)
	return coll
}

// DoRun will realize every element in a lazy sequence.
func DoRun(coll Sequence) {
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
func Slice(coll Sequence) []interface{} {
	return Reduce(func(a []interface{}, b interface{}) []interface{} {
		return append(a, b)
	}, []interface{}{}, coll).([]interface{})
}

// Concat returns a lazy sequence that is the concatenation of the provided
// sequences.
func Concat(colls ...Sequence) Sequence {
	return XfrmSequence(transduce.Cat(Reduce), Seq(colls))
}

// Mapcat returns a lazy sequence that is the concatenation of the
// provided sequences modified by the mapping function f.
// f must be of the form func(in iT) oT and will be called with
// reflection unless it is the non-specialized func(interface{})interface{}.
func Mapcat(f interface{}, colls ...Sequence) Sequence {
	return XfrmSequence(transduce.Mapcat(Reduce, f), Seq(colls))
}

// PartitionBy returns a lazy sequence that consists of partitions of
// the provided sequence. The partitions are determined by f which is
// any function of type func(i iT) oT. When f returns a different value
// from its previous call then a parition is created.
func PartitionBy(f interface{}, coll Sequence) Sequence {
	return XfrmSequence(
		transduce.Compose(transduce.PartitionBy(f),
			transduce.Map(Seq)),
		Seq(coll))
}

// PartitionAll returns a lazy sequence that consists of partitions of
// the provided sequence of n elements. If the length of the sequence
// is not a multiple of n then the remainder will be returned as the
// last element.
func PartitionAll(n int, coll Sequence) Sequence {
	return XfrmSequence(
		transduce.Compose(transduce.PartitionAll(n),
			transduce.Map(Seq)),
		Seq(coll))
}

// apply allows one to call arbitrary go functions using reflection.
// It handles the pitfalls of calling functions using reflection
// so that a simple interface is provided to callers.
// This probably belongs in another library of reflection helpers
// but until that is written it will exist here.
func apply(f interface{}, args ...interface{}) interface{} {
	fnv := reflect.ValueOf(f)
	fnt := fnv.Type()
	argvs := make([]reflect.Value, len(args))
	for i, arg := range args {
		if arg == nil {
			fnint := fnt.In(i)
			fnink := fnint.Kind()
			switch fnink {
			case reflect.Chan, reflect.Func,
				reflect.Interface, reflect.Map,
				reflect.Ptr, reflect.Slice:
				argvs[i] = reflect.Zero(fnint)
			default:
				// this will cause a panic but that is what is
				// intended
				argvs[i] = reflect.ValueOf(arg)
			}
		} else {
			argvs[i] = reflect.ValueOf(arg)
		}
	}
	outvs := fnv.Call(argvs)
	switch len(outvs) {
	case 0:
		return nil
	case 1:
		return outvs[0].Interface()
	default:
		outs := make([]interface{}, len(outvs))
		for i, outv := range outvs {
			outs[i] = outv.Interface()
		}
		return outs
	}
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
