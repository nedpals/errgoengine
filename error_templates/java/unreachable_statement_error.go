package java

import (
	lib "github.com/nedpals/errgoengine"
)

var UnreachableStatementError = lib.ErrorTemplate{
	Name:              "UnreachableStatementError",
	Pattern:           comptimeErrorPattern("unreachable statement"),
	StackTracePattern: comptimeStackTracePattern,
	OnAnalyzeErrorFn: func(cd *lib.ContextData, m *lib.MainError) {

	},
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		gen.Add("This error occurs because there's code after a return statement, which can never be reached as the function has already exited.")
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		gen.Add("Remove unreachable code", func(s *lib.BugFixSuggestion) {
			startPos := cd.MainError.Nearest.StartPosition()
			endPos := cd.MainError.Nearest.Parent().LastNamedChild().EndPosition()

			// Adjust the start position to the beginning of the line
			startPos = startPos.Add(lib.Position{Column: -startPos.Column})

			s.AddStep(
				"Since the `return` statement is encountered before `%s`, the latter statement is unreachable. Remove the unreachable statement/s.",
				cd.MainError.Nearest.Text(),
			).AddFix(lib.FixSuggestion{
				NewText:       "",
				StartPosition: startPos,
				EndPosition:   endPos,
			})
		})
	},
}
