package java

import lib "github.com/nedpals/errgoengine"

var UnreachableStatementError = lib.ErrorTemplate{
	Name:              "UnreachableStatementError",
	Pattern:           comptimeErrorPattern("unreachable statement"),
	StackTracePattern: comptimeStackTracePattern,
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		// TODO: identify return
		gen.Add("You have code below after you returned a value")
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		// TODO:
	},
}
