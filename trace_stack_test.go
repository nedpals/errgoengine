package errgoengine

import (
	"testing"

	testutils "github.com/nedpals/errgoengine/test_utils"
)

func TestTraceStack(t *testing.T) {
	stack := TraceStack{}
	stack.Add("a", Location{
		DocumentPath: "a",
		StartPos:     Position{1, 1, 0},
		EndPos:       Position{1, 1, 0},
	})

	testutils.Equals(t, stack.Top(), StackTraceEntry{
		SymbolName: "a",
		Location: Location{
			DocumentPath: "a",
			StartPos:     Position{1, 1, 0},
			EndPos:       Position{1, 1, 0},
		},
	})

	stack.Add("bbb", Location{
		DocumentPath: "ab/aa",
		StartPos:     Position{1, 1, 0},
		EndPos:       Position{1, 1, 0},
	})

	testutils.Equals(t, stack.Top(), StackTraceEntry{
		SymbolName: "bbb",
		Location: Location{
			DocumentPath: "ab/aa",
			StartPos:     Position{1, 1, 0},
			EndPos:       Position{1, 1, 0},
		},
	})

	stack.Add("cc", Location{
		DocumentPath: "ab/cc",
		StartPos:     Position{1, 1, 0},
		EndPos:       Position{1, 1, 0},
	})

	testutils.Equals(t, stack.NearestTo("ab/aa"), StackTraceEntry{
		SymbolName: "bbb",
		Location: Location{
			DocumentPath: "ab/aa",
			StartPos:     Position{1, 1, 0},
			EndPos:       Position{1, 1, 0},
		},
	})
}
