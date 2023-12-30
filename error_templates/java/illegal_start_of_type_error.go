package java

import lib "github.com/nedpals/errgoengine"

var IllegalStartOfTypeError = lib.ErrorTemplate{
	Name:              "IllegalStartOfTypeError",
	Pattern:           comptimeErrorPattern(`illegal start of type`),
	StackTracePattern: comptimeStackTracePattern,
	OnAnalyzeErrorFn: func(cd *lib.ContextData, m *lib.MainError) {
	},
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		gen.Add("This error occurs when there is an illegal start of a type, typically due to a misplaced return statement.")
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		gen.Add("Test", func(s *lib.BugFixSuggestion) {})
	},
}
