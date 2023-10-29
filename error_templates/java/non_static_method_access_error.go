package java

import lib "github.com/nedpals/errgoengine"

var NonStaticMethodAccessError = lib.ErrorTemplate{
	Name:              "NonStaticMethodAccessError",
	Pattern:           comptimeErrorPattern(`non-static method (?P<method>\S+)\(\) cannot be referenced from a static context`),
	StackTracePattern: comptimeStackTracePattern,
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		// return "TODO:"
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		// TODO
	},
}
