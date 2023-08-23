package java

import (
	lib "github.com/nedpals/errgoengine"
)

var ParseEndOfFileError = lib.ErrorTemplate{
	Name:              "ParseEndOfFileError",
	Pattern:           comptimeErrorPattern("reached end of file while parsing"),
	StackTracePattern: comptimeStackTracePattern,
	OnGenExplainFn: func(cd *lib.ContextData) string {
		// TODO:
		panic("TODO!")
	},
	OnGenBugFixFn: func(cd *lib.ContextData) []lib.BugFix {
		// TODO:
		return make([]lib.BugFix, 0)
	},
}
