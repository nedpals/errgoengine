package java

import (
	lib "github.com/nedpals/errgoengine"
)

var UnknownVariableError = lib.ErrorTemplate{
	Name:              "UnknownVariableError",
	Pattern:           comptimeErrorPattern("cannot find symbol", `symbol:\s+variable (?P<variable>\S+)`),
	StackTracePattern: comptimeStackTracePattern,
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		gen.Add(`The program cannot find variable "%s"`, cd.Variables["variable"])
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		// TODO:
	},
}
