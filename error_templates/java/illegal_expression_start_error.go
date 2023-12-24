package java

import (
	"fmt"
	"strings"

	lib "github.com/nedpals/errgoengine"
)

var IllegalExpressionStartError = lib.ErrorTemplate{
	Name:              "IllegalExpressionStartError",
	Pattern:           comptimeErrorPattern(`illegal start of expression`),
	StackTracePattern: comptimeStackTracePattern,
	OnAnalyzeErrorFn: func(cd *lib.ContextData, m *lib.MainError) {
		lib.QueryNode(m.Nearest, strings.NewReader("(ERROR) @error"), func(ctx lib.QueryNodeCtx) bool {
			match := ctx.Cursor.FilterPredicates(ctx.Match, []byte(m.Nearest.Doc.Contents))
			for _, c := range match.Captures {
				node := lib.WrapNode(m.Nearest.Doc, c.Node)
				m.Nearest = node
				// aCtx.NearestClass = node
				return false
			}
			return true
		})
	},
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		gen.Add("This error occurs when the compiler encounters an expression that is not valid.")
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		parent := cd.MainError.Nearest.Parent()

		if parent.Type() == "unary_expression" {
			firstChild := parent.Child(0)
			operand := parent.ChildByFieldName("operand")
			fmt.Println(firstChild.Text())

			gen.Add("Correct the expression", func(s *lib.BugFixSuggestion) {
				s.AddStep("Ensure a valid expression by fixing the incorrect use of operators").
					AddFix(lib.FixSuggestion{
						NewText:       operand.Text(),
						StartPosition: firstChild.StartPosition(),
						EndPosition:   firstChild.EndPosition(),
					}).
					AddFix(lib.FixSuggestion{
						NewText:       "(",
						StartPosition: parent.StartPosition(),
						EndPosition:   parent.StartPosition(),
					}).
					AddFix(lib.FixSuggestion{
						NewText:       ")",
						StartPosition: parent.EndPosition(),
						EndPosition:   parent.EndPosition(),
					})
			})
		}
	},
}
