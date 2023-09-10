package java

import lib "github.com/nedpals/errgoengine"

var UnclosedCharacterLiteralError = lib.ErrorTemplate{
	Name:              "UnclosedCharacterLiteralError",
	Pattern:           comptimeErrorPattern(`unclosed character literal`),
	StackTracePattern: comptimeStackTracePattern,
	OnGenExplainFn: func(cd *lib.ContextData) string {
		return "Unclosed character literal"
	},
	OnGenBugFixFn: func(cd *lib.ContextData) []lib.BugFix {
		// TODO:
		return make([]lib.BugFix, 0)
	},
}
