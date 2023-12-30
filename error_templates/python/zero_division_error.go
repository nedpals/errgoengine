package python

import (
	"fmt"

	lib "github.com/nedpals/errgoengine"
)

var ZeroDivisionError = lib.ErrorTemplate{
	Name:    "ZeroDivisionError",
	Pattern: "ZeroDivisionError: division by zero",
	OnAnalyzeErrorFn: func(cd *lib.ContextData, m *lib.MainError) {
		for q := m.Nearest.Query("(binary_operator right: (_) @right)"); q.Next(); {
			m.Nearest = q.CurrentNode()
			break
		}
	},
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		// TODO:
		gen.Add("This error occurs when there is an attempt to divide a number by zero.")
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		gen.Add("Avoid division by zero", func(s *lib.BugFixSuggestion) {
			recommendedNumber := 2

			s.AddStep("Ensure that the denominator in a division operation is not zero.").
				AddFix(lib.FixSuggestion{
					NewText:       fmt.Sprintf("%d", recommendedNumber),
					StartPosition: cd.MainError.Nearest.StartPosition(),
					EndPosition:   cd.MainError.Nearest.EndPosition(),
				})
		})
	},
}
