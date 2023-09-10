package java

import lib "github.com/nedpals/errgoengine"

var NonStaticMethodAccessError = lib.ErrorTemplate{
	Name:              "NonStaticMethodAccessError",
	Pattern:           comptimeErrorPattern(`non-static method (?P<method>\S+)\(\) cannot be referenced from a static context`),
	StackTracePattern: comptimeStackTracePattern,
	OnGenExplainFn: func(cd *lib.ContextData) string {
		return "TODO:"
	},
	OnGenBugFixFn: func(cd *lib.ContextData) []lib.BugFix {
		// TODO:
		return make([]lib.BugFix, 0)
	},
}
