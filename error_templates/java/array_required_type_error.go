package java

import "github.com/nedpals/errgoengine/lib"

var ArrayRequiredTypeError = lib.ErrorTemplate{
	Name:              "ArrayRequiredTypeError",
	Pattern:           comptimeErrorPattern(`array required, but (?P<foundType>\S+) found`),
	StackTracePattern: comptimeStackTracePattern,
	OnGenExplainFn: func(cd *lib.ContextData) string {
		// TODO:
		panic("array required type error TODO")
	},
	OnGenBugFixFn: func(cd *lib.ContextData) []lib.BugFix {
		return make([]lib.BugFix, 0)
	},
}
