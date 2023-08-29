package java

import lib "github.com/nedpals/errgoengine"

var UnreachableStatementError = lib.ErrorTemplate{
	Name:              "UnreachableStatementError",
	Pattern:           comptimeErrorPattern("unreachable statement"),
	StackTracePattern: comptimeStackTracePattern,
	OnGenExplainFn: func(cd *lib.ContextData) string {
		// TODO: identify return
		return "You have code below after you returned a value"
	},
	OnGenBugFixFn: func(cd *lib.ContextData) []lib.BugFix {
		return make([]lib.BugFix, 0)
	},
}
