package java

import (
	"strings"

	lib "github.com/nedpals/errgoengine"
)

var IllegalExpressionStartError = lib.ErrorTemplate{
	Name:              "IllegalExpressionStartError",
	Pattern:           comptimeErrorPattern(`illegal start of expression`),
	StackTracePattern: comptimeStackTracePattern,
	OnAnalyzeErrorFn: func(cd *lib.ContextData, m *lib.MainError) {
		for nearest := m.Nearest; !nearest.IsNull(); nearest = nearest.Parent() {
			found := false

			for q := nearest.Query("(ERROR) @error"); q.Next(); {
				node := q.CurrentNode()
				m.Nearest = node
				found = true
				// aCtx.NearestClass = node
				break
			}

			if found {
				break
			}
		}
	},
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		gen.Add("This error occurs when the compiler encounters an expression that is not valid.")
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		parent := cd.MainError.Nearest.Parent()

		if parent.Type() == "unary_expression" {
			firstChild := parent.Child(0)
			operand := parent.ChildByFieldName("operand")

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
		} else if errNodeText := cd.MainError.Nearest.Text(); !strings.HasPrefix(errNodeText, "}") && strings.HasSuffix(errNodeText, "else") {
			gen.Add("Use the right closing bracket", func(s *lib.BugFixSuggestion) {
				s.AddStep("Ensure that the right closing bracket for the else branch of your if statement is used").
					AddFix(lib.FixSuggestion{
						NewText:       "} else",
						StartPosition: cd.MainError.Nearest.StartPosition(),
						EndPosition:   cd.MainError.Nearest.EndPosition(),
					})
			})
		}
	},
}
