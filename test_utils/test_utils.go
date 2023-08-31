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

func ExpectError(tb testing.TB, err error, exp string) {
	if err == nil {
		tb.Fatalf("expected error, got nil")
	}

	Equals(tb, err.Error(), exp)
}

func ExpectNil(tb testing.TB, d any) {
	Equals(tb, fmt.Sprintf("%v", d), "<nil>")
}
