package java

import lib "github.com/nedpals/errgoengine"

var UnclosedCharacterLiteralError = lib.ErrorTemplate{
	Name:              "UnclosedCharacterLiteralError",
	Pattern:           comptimeErrorPattern(`unclosed character literal`),
	StackTracePattern: comptimeStackTracePattern,
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		gen.Add("Unclosed character literal")
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		// TODO:
	},
}
