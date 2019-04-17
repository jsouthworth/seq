package seq

import (
	"testing"
	"testing/quick"
)

func TestReflectIntSlice(t *testing.T) {
	if err := quick.Check(func(is []int) bool {
		got := Seq(is)
		for _, v := range is {
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

func TestReflectStringSlice(t *testing.T) {
	if err := quick.Check(func(is []string) bool {
		got := Seq(is)
		for _, v := range is {
			gv := First(got).(string)
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

func TestReflectString(t *testing.T) {
	if err := quick.Check(func(str string) bool {
		got := Seq(str)
		for _, v := range str {
			gv := First(got).(rune)
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

func TestReflectMap(t *testing.T) {
	if err := quick.Check(func(m map[string]string) bool {
		got := Seq(m)
		for got != nil {
			gv := First(got).(MapEntry)
			v, ok := m[gv.Key().(string)]
			if !ok {
				return false
			}
			if gv.Value() != v {
				t.Logf("%q %q\n", gv.Value(), v)
				return false
			}
			got = Next(got)
		}
		return true
	}, nil); err != nil {
		t.Error(err)
	}
}
func TestReflectMapTraversesAll(t *testing.T) {
	m := map[int]int{
		1: 1,
		2: 2,
		3: 3,
		4: 4,
		5: 5,
	}
	v := Reduce(func(a int, b MapEntry) int {
		return a + b.Value().(int)
	}, 0, m)
	if v != 1+2+3+4+5 {
		t.Fatal("didn't get expected result")
	}

}
