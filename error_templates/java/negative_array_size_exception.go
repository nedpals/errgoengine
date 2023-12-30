package java

import (
	lib "github.com/nedpals/errgoengine"
)

type negativeArraySizeExceptionCtx struct {
	ArrayExprNode lib.SyntaxNode
}

var NegativeArraySizeException = lib.ErrorTemplate{
	Name:    "NegativeArraySizeException",
	Pattern: runtimeErrorPattern("java.lang.NegativeArraySizeException", "(?P<index>.+)"),
	OnAnalyzeErrorFn: func(cd *lib.ContextData, m *lib.MainError) {
		nCtx := negativeArraySizeExceptionCtx{}
		query := "(array_creation_expression dimensions: (dimensions_expr (unary_expression operand: (decimal_integer_literal)))) @array"
		for q := m.Nearest.Query(query); q.Next(); {
			node := q.CurrentNode()
			nCtx.ArrayExprNode = node
			m.Nearest = node.ChildByFieldName("dimensions").NamedChild(0)
			break
		}

		m.Context = nCtx
	},
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		gen.Add("This error occurs when you try to create an array with a negative size.")
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		gen.Add("Ensure a non-negative array size", func(s *lib.BugFixSuggestion) {
			s.AddStep("Make sure the array size is non-negative").
				AddFix(lib.FixSuggestion{
					NewText:       cd.MainError.Nearest.NamedChild(0).Text(),
					StartPosition: cd.MainError.Nearest.StartPosition(),
					EndPosition:   cd.MainError.Nearest.EndPosition(),
				})
		})
	},
}
