package java

import "github.com/nedpals/errgoengine/lib"

var UnreachableStatementError = lib.ErrorTemplate{
	Name:              "UnreachableStatementError",
	Pattern:           comptimeErrorPattern("unreachable statement"),
	StackTracePattern: comptimeStackTracePattern,
	OnGenExplainFn: func(cd *lib.ContextData) string {
		// TODO:
		panic("unreachable statement TODO!")
	},
	OnGenBugFixFn: func(cd *lib.ContextData) []lib.BugFix {
		return make([]lib.BugFix, 0)
	},
}