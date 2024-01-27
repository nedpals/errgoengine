package java

import (
	lib "github.com/nedpals/errgoengine"
)

var IllegalCharacterError = lib.ErrorTemplate{
	Name:              "IllegalCharacterError",
	Pattern:           comptimeErrorPattern(`illegal character: '(?P<character>\S+)'`),
	StackTracePattern: comptimeStackTracePattern,
	OnAnalyzeErrorFn: func(cd *lib.ContextData, m *lib.MainError) {
		for q := m.Nearest.Query("(ERROR) @error"); q.Next(); {
			node := q.CurrentNode()
			m.Nearest = node.NextSibling()
			break
		}
	},
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		gen.Add("This error occurs when there is an attempt to use an illegal character in the code.")
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		parent := cd.MainError.Nearest.Parent()

		if parent.Type() == "binary_expression" {
			gen.Add("Remove the illegal character", func(s *lib.BugFixSuggestion) {
				left := parent.ChildByFieldName("left")

				s.AddStep("Remove the illegal character `%s` from the code", cd.Variables["character"]).
					AddFix(lib.FixSuggestion{
						NewText:       left.Text(),
						StartPosition: parent.StartPosition(),
						EndPosition:   parent.EndPosition(),
					})
			})
		}
	},
}
