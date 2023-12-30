package java

import lib "github.com/nedpals/errgoengine"

var ExpectedError = lib.ErrorTemplate{
	Name:              "ExpectedError",
	Pattern:           comptimeErrorPattern(`(?P<expected>.+) expected`),
	StackTracePattern: comptimeStackTracePattern,
	OnAnalyzeErrorFn: func(cd *lib.ContextData, m *lib.MainError) {
	},
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
	},
}
