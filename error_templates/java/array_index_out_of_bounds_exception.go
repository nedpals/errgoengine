package java

import (
	lib "github.com/nedpals/errgoengine"
)

var ArrayIndexOutOfBoundsException = lib.ErrorTemplate{
	Name:    "ArrayIndexOutOfBoundsException",
	Pattern: runtimeErrorPattern("java.lang.ArrayIndexOutOfBoundsException", `Index (?P<index>\d+) out of bounds for length (?P<length>\d+)`),
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		// TODO:
		gen.Add("Your program attempted to access an element in index %s on an array that only has %s items", cd.Variables["index"], cd.Variables["length"])
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		// TODO:
	},
}
