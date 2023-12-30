package java

import (
	"fmt"
	"strconv"

	lib "github.com/nedpals/errgoengine"
)

var ArrayIndexOutOfBoundsException = lib.ErrorTemplate{
	Name:    "ArrayIndexOutOfBoundsException",
	Pattern: runtimeErrorPattern("java.lang.ArrayIndexOutOfBoundsException", `Index (?P<index>\d+) out of bounds for length (?P<length>\d+)`),
	OnAnalyzeErrorFn: func(cd *lib.ContextData, m *lib.MainError) {
		for q := m.Nearest.Query(`(array_access index: (_) @index)`); q.Next(); {
			node := q.CurrentNode()
			m.Nearest = node
			break
		}
	},
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		gen.Add("This error occurs because the code is trying to access index %s that is beyond the bounds of the array which only has %s items.", cd.Variables["index"], cd.Variables["length"])
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		arrayLen, _ := strconv.Atoi(cd.Variables["length"])

		// TODO: add a suggestion to add an if statement if the array length is 0

		gen.Add("Accessing Array Index Within Bounds", func(s *lib.BugFixSuggestion) {
			sampleIndex := max(0, arrayLen-2)

			s.AddStep("The error is caused by trying to access an index that does not exist within the array. Instead of accessing index %s, which is beyond the array's length, change it to a valid index within the array bounds, for example, `nums[%d]`.", cd.Variables["index"], sampleIndex).
				AddFix(lib.FixSuggestion{
					NewText:       fmt.Sprintf("%d", sampleIndex),
					StartPosition: cd.MainError.Nearest.StartPosition(),
					EndPosition:   cd.MainError.Nearest.EndPosition(),
					Description:   "This adjustment ensures that you're accessing an index that exists within the array bounds, preventing the `ArrayIndexOutOfBoundsException`.",
				})
		})
	},
}
