package python

import (
	lib "github.com/nedpals/errgoengine"
)

var NameError = lib.ErrorTemplate{
	Name:    "NameError",
	Pattern: `NameError: name '(?P<variable>\S+)' is not defined`,
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		gen.Add("Your program tried to access the '%s' variable which was not found on your program.", cd.Variables["variable"])
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		// TODO:
	},
}
