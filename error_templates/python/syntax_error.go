package python

import lib "github.com/nedpals/errgoengine"

var SyntaxError = lib.ErrorTemplate{
	Name:    "SyntaxError",
	Pattern: compileTimeError("SyntaxError: (?P<reason>.+)"),
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		// TODO:
		if cd.Variables["reason"] == "'(' was never closed" {
			gen.Add("Your program did not close a parenthesis properly")
			return
		}
		gen.Add("The interpreter was not able to understand your program because of an invalid syntax")
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		// TODO
	},
}
