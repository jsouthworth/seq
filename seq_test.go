package seq

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"

	"jsouthworth.net/go/transduce"
)

func TestFirst(t *testing.T) {
	if err := quick.Check(func(is []int) bool {
		return len(is) < 1 || First(Seq(is)) == is[0]
	}, nil); err != nil {
		t.Error(err)
	}
}

func TestConjSlice(t *testing.T) {
	if err := quick.Check(func(is []int, other int) bool {
		new := Conj(is, other).([]int)
		return len(new) == len(is)+1 &&
			new[len(new)-1] == other
	}, nil); err != nil {
		t.Error(err)
	}
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

func TestConjMap(t *testing.T) {
	if err := quick.Check(func(is map[string]int, otherk string, other int) bool {
		new := Conj(is, mapEntry{otherk, other}).(map[string]int)
		return new[otherk] == other
	}, nil); err != nil {
		t.Error(err)
	}
}

type intSliceConjoiner struct {
	slice []int
}

func (c *intSliceConjoiner) Conj(elem interface{}) interface{} {
	c.slice = append(c.slice, elem.(int))
	return c.slice
}

func TestConjConjoiner(t *testing.T) {
	if err := quick.Check(func(is []int, other int) bool {
		new := Conj(&intSliceConjoiner{is}, other).([]int)
		return len(new) == len(is)+1 &&
			new[len(new)-1] == other
	}, nil); err != nil {
		t.Error(err)
	}
}

func TestInvalidConj(t *testing.T) {
	if err := quick.Check(func(i int, other int) (out bool) {
		defer func() {
			r := recover()
			if r != nil {
				out = true
			}
		}()
		_ = Conj(i, other)
		return false
	}, nil); err != nil {
		t.Error(err)
	}
}

func TestInfo(t *testing.T) {
	if err := quick.Check(func(is []int) bool {
		new := Into([]int{}, is).([]int)
		var allMatched = true
		for i, v := range new {
			if is[i] != v {
				allMatched = false
			}
		}
		return len(new) == len(is) && allMatched
	}, nil); err != nil {
		t.Error(err)
	}
}

func TestTransformInfo(t *testing.T) {
	if err := quick.Check(func(is []int) bool {
		new := TransformInto([]int{},
			transduce.Map(func(in int) interface{} {
				return in * in
			}),
			is).([]int)
		var allMatched = true
		for i, v := range is {
			if new[i] != v*v {
				allMatched = false
			}
		}
		return len(new) == len(is) && allMatched
	}, nil); err != nil {
		t.Error(err)
	}
}

func ExampleFirst() {
	fmt.Println(First(RangeUntil(10)))
	// Output: 0
}

func TestNext(t *testing.T) {
	if err := quick.Check(func(is []int) bool {
		return len(is) < 2 || First(Next(Seq(is))) == is[1]
	}, nil); err != nil {
		t.Error(err)
	}
}

func ExampleNext() {
	fmt.Println(Next(RangeUntil(10)))
	// Output: (1 2 3 4 5 6 7 8 9)
}

func TestMap(t *testing.T) {
	if err := quick.Check(func(is []int) bool {
		expected := make([]int, len(is))
		for i, v := range is {
			expected[i] = v + v
		}
		got := Seq(Map(func(a interface{}) interface{} {
			return a.(int) + a.(int)
		}, Seq(is)))
		for _, v := range expected {
			gv := First(got).(int)
			if gv != v {
				return false
			}
			got = Next(got)
		}
		return true
	}, nil); err != nil {
		t.Error(err)
	}
}

func TestMapReflect(t *testing.T) {
	if err := quick.Check(func(is []int) bool {
		expected := make([]int, len(is))
		for i, v := range is {
			expected[i] = v + v
		}
		got := Seq(Map(func(a int) int {
			return a + a
		}, Seq(is)))
		for _, v := range expected {
			gv := First(got).(int)
			if gv != v {
				return false
			}
			got = Next(got)
		}
		return true
	}, nil); err != nil {
		t.Error(err)
	}
}

func ExampleMap() {
	fmt.Println(Map(func(a int) int {
		return a + a
	}, RangeUntil(10)))
	// Output: (0 2 4 6 8 10 12 14 16 18)
}

func ExampleReplace() {
	fmt.Println(Replace(map[int]int{
		1: 10,
	}, RangeUntil(10)))
	// Output: (0 10 2 3 4 5 6 7 8 9)
}

func TestTransduce(t *testing.T) {
	if err := quick.Check(func(is []int) bool {
		expected := make([]int, len(is))
		for i, v := range is {
			expected[len(is)-i-1] = v + v
		}
		got := Seq(Transduce(
			transduce.Map(func(in interface{}) interface{} {
				return in.(int) + in.(int)
			}),
			func(result, input interface{}) interface{} {
				return Cons(input, Seq(result))
			},
			Empty(),
			Seq(is),
		))

		for _, v := range expected {
			gv := First(got).(int)
			if gv != v {
				return false
			}
			got = Next(got)
		}
		return true
	}, nil); err != nil {
		t.Error(err)
	}
}

func TestTransduceReflect(t *testing.T) {
	if err := quick.Check(func(is []int) bool {
		expected := make([]int, len(is))
		for i, v := range is {
			expected[len(is)-i-1] = v + v
		}
		got := Seq(Transduce(
			transduce.Map(func(in int) int {
				return in + in
			}),
			func(result Sequence, input int) Sequence {
				return Cons(input, Seq(result))
			},
			Empty(),
			Seq(is),
		))

		for _, v := range expected {
			gv := First(got).(int)
			if gv != v {
				return false
			}
			got = Next(got)
		}
		return true
	}, nil); err != nil {
		t.Error(err)
	}
}

func ExampleTransduce() {
	fmt.Println(Transduce(
		transduce.Map(func(in int) int {
			return in + in
		}),
		func(result Sequence, input int) Sequence {
			return Cons(input, Seq(result))
		},
		Empty(),
		RangeUntil(10),
	))
	// Output: (18 16 14 12 10 8 6 4 2 0)
}

func TestXfrmSequenceIsLazy(t *testing.T) {
	if err := quick.Check(func(is []int) bool {
		got := Seq(Map(func(a interface{}) interface{} {
			return a.(int) + a.(int)
		}, Seq(is)))
		if got != nil {
			xseq := got.(*xfrmSeq)
			next := xseq.bufferedColl.(*cons).next
			if next != nil {
				nextXfrm := next.(*xfrmSeq)
				return nextXfrm.bufferedColl == nil
			}
		}
		return true
	}, nil); err != nil {
		t.Error(err)
	}
}

func TestReduce(t *testing.T) {
	if err := quick.Check(func(is []int) bool {
		expected := 0
		for _, v := range is {
			expected += v
		}
		got := Reduce(func(a, b interface{}) interface{} {
			return a.(int) + b.(int)
		}, 0, Seq(is))
		return expected == got
	}, nil); err != nil {
		t.Error(err)
	}
}

func TestReduceReflect(t *testing.T) {
	if err := quick.Check(func(is []int) bool {
		expected := 0
		for _, v := range is {
			expected += v
		}
		got := Reduce(func(a, b int) int {
			return a + b
		}, 0, Seq(is))
		return expected == got
	}, nil); err != nil {
		t.Error(err)
	}
}

func ExampleReduce() {
	fmt.Println(Reduce(func(a, b int) int {
		return a + b
	}, 0, RangeUntil(10)))
	// Output: 45
}

func TestSlice(t *testing.T) {
	if err := quick.Check(func(is []int) bool {
		got := Slice(Seq(is))
		for i, v := range is {
			gv := got[i]
			if gv != v {
				return false
			}
		}
		return true
	}, nil); err != nil {
		t.Error(err)
	}
}

func ExampleSlice() {
	fmt.Println(Slice(RangeUntil(10)))
	// Output: [0 1 2 3 4 5 6 7 8 9]
}

func TestRange(t *testing.T) {
	t.Run("step>zero&&start<end", func(t *testing.T) {
		rng := RangeBetween(1, 10)
		expected := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
		seq := rng
		for _, i := range expected {
			f := seq.First()
			if i != f {
				t.Fatal("wanted", i, "got", f)
			}
			seq = seq.Next()
		}
		if First(seq) != nil {
			t.Fatal("unexpected value", First(seq))
		}
	})
	t.Run("step>zero&&start>end", func(t *testing.T) {
		rng := RangeBetween(10, 1)
		if First(rng) != nil {
			t.Fatal("unexpected value", First(rng))
		}
	})
	t.Run("step<zero&&start<end", func(t *testing.T) {
		rng := Range(1, 10, -1)
		if First(rng) != nil {
			t.Fatal("unexpected value", First(rng))
		}
	})
	t.Run("step<zero&&start>end", func(t *testing.T) {
		rng := Range(10, 1, -1)
		expected := []int{10, 9, 8, 7, 6, 5, 4, 3, 2}
		seq := rng
		for _, i := range expected {
			f := seq.First()
			if i != f {
				t.Fatal("wanted", i, "got", f)
			}
			seq = seq.Next()
		}
		if First(seq) != nil {
			t.Fatal("unexpected value", First(seq))
		}
		if Next(seq) != nil {
			t.Fatal("unexpected value", Next(seq))
		}
	})
	t.Run("start==end&&step==0", func(t *testing.T) {
		rng := Range(10, 10, 0)
		if First(rng) != nil {
			t.Fatal("unexpected value", First(rng))
		}
	})
	t.Run("start==end&&step==1", func(t *testing.T) {
		rng := Range(10, 10, 1)
		if First(rng) != nil {
			t.Fatal("unexpected value", First(rng))
		}
	})
	t.Run("start!=end&&step==0", func(t *testing.T) {
		rng := Range(9, 10, 0)
		if rng.First() != 9 {
			t.Fatal("unexpected value", rng.First())
		}
	})
	t.Run("RangeUntil", func(t *testing.T) {
		rng := RangeUntil(10)
		expected := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
		seq := rng
		for _, i := range expected {
			f := seq.First()
			if i != f {
				t.Fatal("wanted", i, "got", f)
			}
			seq = seq.Next()
		}
		if First(seq) != nil {
			t.Fatal("unexpected value", First(seq))
		}
	})
}

func ExampleRange() {
	fmt.Println(Range(1, 10, 2))
	// Output: (1 3 5 7 9)
}

func ExampleRangeBetween() {
	fmt.Println(RangeBetween(1, 5))
	// Output: (1 2 3 4)
}

func ExampleRangeUntil() {
	fmt.Println(RangeUntil(10))
	// Output: (0 1 2 3 4 5 6 7 8 9)
}

func TestRepeat(t *testing.T) {
	t.Run("Repeat", func(t *testing.T) {
		seq := Repeat(10, "foo")
		for i := 0; i < 10; i++ {
			if seq.First() != "foo" {
				t.Fatal("unexpected value", seq.First())
			}
			seq = seq.Next()
		}
		if Next(seq) != nil {
			t.Fatal("unexpected sequence", Next(seq))
		}
	})
	t.Run("RepeateInfinitely", func(t *testing.T) {
		if err := quick.Check(func(n boundedInt) bool {
			seq := RepeateInfinitely("foo")
			for i := 0; i < int(n); i++ {
				if seq.First() != "foo" {
					t.Fatal("unexpected value", seq.First())
				}
				seq = seq.Next()
			}
			return true
		}, nil); err != nil {
			t.Error(err)
		}
	})
}

func ExampleRepeat() {
	fmt.Println(Repeat(10, "foo"))
	// Output: (foo foo foo foo foo foo foo foo foo foo)
}

func ExampleRepeatInfinitely() {
	fmt.Println(Take(10, RepeateInfinitely("foo")))
	// Output: (foo foo foo foo foo foo foo foo foo foo)
}

func TestIterate(t *testing.T) {
	double := func(x int) int {
		return x + x
	}
	sum := func(result, input int) int {
		return result + input
	}
	total := Reduce(sum, 0, Take(10, Iterate(double, 2)))
	if total != 2046 {
		t.Fatal("Iter didn't return expected result")
	}
}

func ExampleIterate() {
	double := func(x int) int {
		return x + x
	}
	fmt.Println(Take(10, Iterate(double, 2)))
	// Output: (2 4 8 16 32 64 128 256 512 1024)
}

func TestCycle(t *testing.T) {
	cyc := Cycle(RangeUntil(10))
	expected := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1}
	seq := cyc
	for idx, i := range expected {
		f := seq.First()
		if i != f {
			t.Fatal("wanted", i, "got", f, "at", idx)
		}
		seq = seq.Next()
	}
	if seq.Next() == nil || seq.First() != 2 {
		t.Fatal("cycle did not continue")
	}
}

func ExampleCycle() {
	fmt.Println(Take(15, Cycle(RangeUntil(10))))
	// Output: (0 1 2 3 4 5 6 7 8 9 0 1 2 3 4)
}

func TestCycleRepeat(t *testing.T) {
	cyc := Cycle(Repeat(1, 10))
	expected := []int{10, 10, 10, 10, 10, 10, 10}
	seq := cyc
	for idx, i := range expected {
		f := seq.First()
		if i != f {
			t.Fatal("wanted", i, "got", f, "at", idx)
		}
		seq = seq.Next()
	}
	if seq.Next() == nil || seq.First() != 10 {
		t.Fatal("cycle did not continue")
	}
}

func TestInterleave(t *testing.T) {
	s1 := Seq([]int{1, 2, 3, 4, 5, 6})
	s2 := Seq([]int{7, 8, 9, 10, 11, 12})
	s3 := Seq([]int{13, 14, 15, 16, 17, 18})
	ilv := Interleave(s1, s2, s3)
	expected := []int{
		1, 7, 13,
		2, 8, 14,
		3, 9, 15,
		4, 10, 16,
		5, 11, 17,
		6, 12, 18}
	seq := ilv
	for idx, i := range expected {
		f := seq.First()
		if i != f {
			t.Fatal("wanted", i, "got", f, "at", idx)
		}
		seq = seq.Next()
	}
}

func TestInterleaveUneven(t *testing.T) {
	s1 := Seq([]int{1, 2, 3, 4, 5, 6})
	s2 := Seq([]int{7, 8, 9, 10, 11, 12})
	s3 := Seq([]int{13, 14, 15, 16, 17})
	ilv := Interleave(s1, s2, s3)
	expected := []int{
		1, 7, 13,
		2, 8, 14,
		3, 9, 15,
		4, 10, 16,
		5, 11, 17}
	seq := ilv
	for idx, i := range expected {
		f := First(seq)
		if i != f {
			t.Fatal("wanted", i, "got", f, "at", idx)
		}
		seq = Next(seq)
	}
	if Next(seq) != nil {
		t.Fatal("unexpected value", seq)
	}
}

func ExampleInterleave() {
	s1 := []int{1, 2, 3, 4, 5, 6}
	s2 := []int{7, 8, 9, 10, 11, 12}
	s3 := []int{13, 14, 15, 16, 17, 18}
	fmt.Println(Interleave(s1, s2, s3))
	// Output: (1 7 13 2 8 14 3 9 15 4 10 16 5 11 17 6 12 18)
}

func TestInterpose(t *testing.T) {
	if err := quick.Check(func(s string, is []int) bool {
		ipos := Interpose(s, Seq(is))
		count := 1
		for seq := ipos; seq != nil; seq = Next(seq) {
			if count%2 == 0 {
				if First(seq) != s {
					return false
				}
			}
		}
		return true
	}, nil); err != nil {
		t.Error(err)
	}
}

func ExampleInterpose() {
	fmt.Println(Interpose("\"-\"", RangeUntil(10)))
	// Output: (0 "-" 1 "-" 2 "-" 3 "-" 4 "-" 5 "-" 6 "-" 7 "-" 8 "-" 9)
}

func TestDrop(t *testing.T) {
	if err := quick.Check(func(is []int) bool {
		if len(is) < 2 {
			return true
		}
		dropped := Slice(Drop(len(is)-1, Seq(is)))
		return len(dropped) == 1
	}, nil); err != nil {
		t.Error(err)
	}
}

func ExampleDrop() {
	fmt.Println(Drop(10, RangeUntil(30)))
	// Output: (10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29)
}

func TestTake(t *testing.T) {
	if err := quick.Check(func(is []int) bool {
		if len(is) < 2 {
			return true
		}
		taken := Slice(Take(len(is)-1, Seq(is)))
		return len(taken) == len(is)-1
	}, nil); err != nil {
		t.Error(err)
	}
}

func ExampleTake() {
	fmt.Println(Take(10, RangeUntil(30)))
	// Output: (0 1 2 3 4 5 6 7 8 9)
}

func ExampleTakeNth() {
	fmt.Println(TakeNth(4, RangeUntil(30)))
	// Output: (3 7 11 15 19 23 27)
}

func ExampleEvery() {
	fmt.Println(Every(func(x int) bool { return x == 10 }, Repeat(100, 10)))
	// Output: true
}

func ExampleSome() {
	fmt.Println(Some(func(x int) bool { return x == 10 },
		Repeat(100, 10)))
	// Output: true
}

func ExampleNotEvery() {
	fmt.Println(NotEvery(func(x int) bool { return x == 10 },
		Repeat(100, 10)))
	// Output: false
}

func ExampleNotAny() {
	fmt.Println(NotAny(func(x int) bool { return x == 10 },
		Repeat(100, 10)))
	// Output: false
}

func ExampleFilter() {
	fmt.Println(Filter(func(x int) bool { return x%2 == 0 },
		RangeUntil(10)))
	// Output: (0 2 4 6 8)
}

func ExampleRemove() {
	fmt.Println(Remove(func(x int) bool { return x%2 == 0 },
		RangeUntil(10)))
	// Output: (1 3 5 7 9)
}

func ExampleConcat() {
	fmt.Println(Concat(Seq([]int{1, 2, 3}), Seq([]int{4, 5, 6})))
	// Output: (1 2 3 4 5 6)
}

func ExampleMapcat() {
	fmt.Println(Mapcat(
		func(x Sequence) Sequence {
			return Remove(func(x int) bool {
				return x%2 == 0
			}, x)
		},
		Seq([]int{1, 2, 3}), Seq([]int{4, 5, 6}), Seq([]int{7, 8, 9})))
	// Output: (1 3 5 7 9)
}

func ExampleTakeWhile() {
	fmt.Println(TakeWhile(func(x int) bool { return x < 9 },
		RangeUntil(20)))
	// Output: (0 1 2 3 4 5 6 7 8)
}

func ExampleDropWhile() {
	fmt.Println(DropWhile(func(x int) bool { return x < 9 },
		RangeUntil(20)))
	// Output: (9 10 11 12 13 14 15 16 17 18 19)
}

func ExampleKeep() {
	ifOdd := func(in int) interface{} {
		if in%2 == 0 {
			return nil
		}
		return in
	}
	fmt.Println(Keep(ifOdd, RangeUntil(10)))
	// Output: (1 3 5 7 9)
}

func ExampleKeepIndexed() {
	ifOdd := func(idx, in int) interface{} {
		if in%2 == 0 || idx > 4 {
			return nil
		}
		return in
	}
	fmt.Println(KeepIndexed(ifOdd, RangeUntil(10)))
	// Output: (1 3)
}

func ExampleDedupe() {
	fmt.Println(Dedupe(Seq([]int{1, 1, 1, 2, 2, 3, 3, 3, 3})))
	// Output: (1 2 3)
}

func ExampleSplitWith() {
	fmt.Println(SplitWith(func(x int) bool { return x < 9 },
		RangeUntil(20)))
	// Output: ((0 1 2 3 4 5 6 7 8) (9 10 11 12 13 14 15 16 17 18 19))
}

func ExampleSplitAt() {
	fmt.Println(SplitAt(9, RangeUntil(20)))
	// Output: ((0 1 2 3 4 5 6 7 8) (9 10 11 12 13 14 15 16 17 18 19))
}

func ExamplePartitionBy() {
	fmt.Println(PartitionBy(func(x int) bool { return x%2 != 0 },
		Seq([]int{1, 1, 1, 2, 2, 3, 3})))
	// Output: ((1 1 1) (2 2) (3 3))
}

func ExamplePartitionAll() {
	fmt.Println(PartitionAll(4,
		Seq([]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})))
	// Output: ((0 1 2 3) (4 5 6 7) (8 9))
}

func TestString(t *testing.T) {
	t.Run("Range", func(t *testing.T) {
		s := RangeUntil(10)
		got := fmt.Sprint(s)
		exp := "(0 1 2 3 4 5 6 7 8 9)"
		if got != exp {
			t.Fatalf("String didn't return expected. got %s expected %s",
				got, exp,
			)
		}
	})
	t.Run("Cycle", func(t *testing.T) {
		s := Take(15, Cycle(RangeUntil(10)))
		got := fmt.Sprint(s)
		exp := "(0 1 2 3 4 5 6 7 8 9 0 1 2 3 4)"
		if got != exp {
			t.Fatalf("String didn't return expected. got %s expected %s",
				got, exp,
			)
		}
	})
	t.Run("Cons", func(t *testing.T) {
		s := Cons(1, Cons(2, Cons(3, nil)))
		got := fmt.Sprint(s)
		exp := "(1 2 3)"
		if got != exp {
			t.Fatalf("String didn't return expected. got %s expected %s",
				got, exp,
			)
		}
	})
	t.Run("Iterate", func(t *testing.T) {
		double := func(x int) int {
			return x + x
		}
		exp := "(2 4 8 16 32 64 128 256 512 1024)"
		got := fmt.Sprint(Take(10, Iterate(double, 2)))
		if got != exp {
			t.Fatalf("String didn't return expected. got %s expected %s",
				got, exp,
			)
		}
	})
	t.Run("Lazy", func(t *testing.T) {
		s1 := Seq([]int{1, 2, 3, 4, 5, 6})
		s2 := Seq([]int{7, 8, 9, 10, 11, 12})
		s3 := Seq([]int{13, 14, 15, 16, 17, 18})
		ilv := Interleave(s1, s2, s3)
		exp := "(1 7 13 2 8 14 3 9 15 4 10 16 5 11 17 6 12 18)"
		got := fmt.Sprint(ilv)
		if got != exp {
			t.Fatalf("String didn't return expected. got %s expected %s",
				got, exp,
			)
		}
	})
	t.Run("Repeat", func(t *testing.T) {
		rep := Repeat(2, 10)
		exp := "(10 10)"
		got := fmt.Sprint(rep)
		if got != exp {
			t.Fatalf("String didn't return expected. got %s expected %s",
				got, exp,
			)
		}
	})
	t.Run("Reflect", func(t *testing.T) {
		s := Seq([]int{1, 2, 3, 4, 5, 6})
		exp := "(1 2 3 4 5 6)"
		got := fmt.Sprint(s)
		if got != exp {
			t.Fatalf("String didn't return expected. got %s expected %s",
				got, exp,
			)
		}
	})
}

type boundedInt int

func (i boundedInt) Generate(rand *rand.Rand, size int) reflect.Value {
	v := rand.Uint64()
	return reflect.ValueOf(boundedInt(v % 1000))
}

type IntSeq []int

func (i IntSeq) First() interface{} {
	return i[0]
}
func (i IntSeq) Next() Sequence {
	if len(i) <= 1 {
		return nil
	}
	return i[1:]
}
func BenchmarkMap(b *testing.B) {
	b.Run("native-loop", func(b *testing.B) {
		s := make([]int, b.N)
		for i, v := range s {
			s[i] = v + 10
		}
	})
	b.Run("map-reflect", func(b *testing.B) {
		s := make([]int, b.N)
		seq := Map(func(in interface{}) interface{} {
			return in.(int) + 10
		}, Seq(s))
		DoRun(Seq(seq))
	})
	b.Run("map-reflect-reflected-func", func(b *testing.B) {
		s := make([]int, b.N)
		seq := Map(func(in int) int {
			return in + 10
		}, Seq(s))
		DoRun(Seq(seq))
	})
	b.Run("map-intseq", func(b *testing.B) {
		s := make(IntSeq, b.N)
		seq := Map(func(in interface{}) interface{} {
			return in.(int) + 10
		}, Seq(s))
		DoRun(Seq(seq))
	})
	b.Run("map-intseq-reflected-func", func(b *testing.B) {
		s := make(IntSeq, b.N)
		seq := Map(func(in int) int {
			return in + 10
		}, Seq(s))
		DoRun(Seq(seq))
	})

}

func ExampleXfrmSequence() {
	xform := transduce.Compose(
		transduce.Map(func(x int) int {
			return x + 1
		}),
		transduce.Filter(func(x int) bool {
			return x%2 == 0
		}),
		transduce.Dedupe(),
		transduce.Mapcat(Reduce, RangeUntil),
		transduce.PartitionAll(3),
		transduce.PartitionBy(func(coll interface{}) bool {
			return Reduce(func(res, x int) int { return res + x },
				0, coll).(int) > 7
		}),
		transduce.Cat(Reduce), //TODO: implement flatten and combine
		transduce.Cat(Reduce), //TODO: implement flatten and combine
		transduce.RandomSample(1.0),
		transduce.TakeNth(1),
		transduce.Keep(func(x int) interface{} {
			if x%2 != 0 {
				return x * x
			}
			return nil
		}),
		transduce.KeepIndexed(func(i, x int) interface{} {
			if i%2 == 0 {
				return i * x
			}
			return nil
		}),
		transduce.Replace(map[int]string{
			2:  "two",
			6:  "six",
			18: "eighteen",
		}),
		transduce.Take(11),
		transduce.TakeWhile(func(x interface{}) bool {
			return x != 300
		}),
		transduce.Drop(1),
		transduce.DropWhile(func(x interface{}) bool {
			_, isString := x.(string)
			return isString
		}),
		transduce.Remove(func(x interface{}) bool {
			_, isString := x.(string)
			return isString
		}),
	)
	data := Interleave(RangeUntil(18), RangeUntil(20))
	fmt.Println(XfrmSequence(xform, data))
	// Output: (36 200 10)
}
