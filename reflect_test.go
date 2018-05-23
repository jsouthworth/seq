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
