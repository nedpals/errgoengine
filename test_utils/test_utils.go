package testutils

import (
	"fmt"
	"testing"
)

func Equals[V comparable](tb testing.TB, got V, exp V) {
	if got != exp {
		tb.Fatalf("\nexp: %v\ngot: %v", exp, got)
	}
}

func EqualsMap[K comparable, V comparable](tb testing.TB, got map[K]V, exp map[K]V) {
	if len(got) != len(exp) {
		tb.Fatalf("\nexp (len %d): %v\ngot (len %d): %v", len(exp), exp, len(got), got)
	}

	for k, v := range exp {
		if gotV, gotOk := got[k]; !gotOk {
			tb.Fatalf("\nkey `%v` in got map not present", k)
		} else if v != gotV {
			tb.Fatalf("\nin got[%v]\nexp: %v\ngot: %v", k, v, gotV)
		}
	}
}

func EqualsList[V comparable](tb testing.TB, got []V, exp []V) {
	if len(got) != len(exp) {
		tb.Fatalf("\nexp (len %d): %v\ngot (len %d): %v", len(exp), exp, len(got), got)
	}

	for i, v := range exp {
		if got[i] != v {
			tb.Fatalf("\nin got[%v]\nexp: %v\ngot: %v", i, v, got[i])
		}
	}
}

func ExpectError(tb testing.TB, err error, exp string) {
	if err == nil {
		tb.Fatalf("expected error, got nil")
	}

	Equals(tb, err.Error(), exp)
}

func ExpectNil(tb testing.TB, d any) {
	Equals(tb, fmt.Sprintf("%v", d), "<nil>")
}

func ExpectNoError(tb testing.TB, err error) {
	if err != nil {
		tb.Fatalf("expected none, got error\nerror: %s", err)
	}
}
