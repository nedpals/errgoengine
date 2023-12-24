package java

import (
	"strings"

	lib "github.com/nedpals/errgoengine"
)

var UnclosedStringLiteralError = lib.ErrorTemplate{
	Name:              "UnclosedStringLiteralError",
	Pattern:           comptimeErrorPattern(`unclosed string literal`),
	StackTracePattern: comptimeStackTracePattern,
	OnAnalyzeErrorFn: func(cd *lib.ContextData, m *lib.MainError) {
		// go back to parent if pointed nearest node is not an error node
		if !m.Nearest.IsError() {
			m.Nearest = m.Nearest.Parent()
		}
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
		gen.Add("This error occurs when there is an unclosed string literal in the code.")
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		parent := cd.MainError.Nearest.Parent()

		if parent.Type() == "variable_declarator" {
			gen.Add("Close the string literal", func(s *lib.BugFixSuggestion) {
				s.AddStep("Ensure that the string literal is properly closed with a double-quote.").
					AddFix(lib.FixSuggestion{
						NewText:       "\"",
						StartPosition: cd.MainError.Nearest.EndPosition().Add(lib.Position{Column: -1}),
						EndPosition:   cd.MainError.Nearest.EndPosition().Add(lib.Position{Column: -1}),
					})
			})
		}
	},
}
