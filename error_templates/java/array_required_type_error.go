package java

import (
	lib "github.com/nedpals/errgoengine"
)

var ArrayRequiredTypeError = lib.ErrorTemplate{
	Name:              "ArrayRequiredTypeError",
	Pattern:           comptimeErrorPattern(`array required, but (?P<foundType>\S+) found`),
	StackTracePattern: comptimeStackTracePattern,
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		// TODO:
		gen.Add("You are calling an index notation on a variable with type %s", cd.Variables["foundType"])
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		// TODO:
	},
}
