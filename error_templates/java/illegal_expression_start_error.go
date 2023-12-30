package java

import (
	lib "github.com/nedpals/errgoengine"
)

var IllegalExpressionStartError = lib.ErrorTemplate{
	Name:              "IllegalExpressionStartError",
	Pattern:           comptimeErrorPattern(`illegal start of expression`),
	StackTracePattern: comptimeStackTracePattern,
	OnAnalyzeErrorFn: func(cd *lib.ContextData, m *lib.MainError) {
		for q := m.Nearest.Query("(ERROR) @error"); q.Next(); {
			node := q.CurrentNode()
			m.Nearest = node
			// aCtx.NearestClass = node
			break
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
		}
	},
}
