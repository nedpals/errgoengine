package java

import (
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

		for q := m.Nearest.Query(`(ERROR) @error`); q.Next(); {
			node := q.CurrentNode()
			m.Nearest = node
			// aCtx.NearestClass = node
			break
		}
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
