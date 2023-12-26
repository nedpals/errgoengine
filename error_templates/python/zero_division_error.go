package python

import (
	"fmt"
	"strings"

	lib "github.com/nedpals/errgoengine"
)

var ZeroDivisionError = lib.ErrorTemplate{
	Name:    "ZeroDivisionError",
	Pattern: "ZeroDivisionError: division by zero",
	OnAnalyzeErrorFn: func(cd *lib.ContextData, m *lib.MainError) {
		lib.QueryNode(m.Nearest, strings.NewReader(`(binary_operator right: (_) @right)`), func(ctx lib.QueryNodeCtx) bool {
			match := ctx.Cursor.FilterPredicates(ctx.Match, []byte(m.Nearest.Doc.Contents))
			for _, c := range match.Captures {
				node := lib.WrapNode(m.Nearest.Doc, c.Node)
				m.Nearest = node
				return false
			}
			return true
		})
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
