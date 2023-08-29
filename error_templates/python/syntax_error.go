package python

import lib "github.com/nedpals/errgoengine"

var SyntaxError = lib.ErrorTemplate{
	Name:    "SyntaxError",
	Pattern: compileTimeError("SyntaxError: (?P<reason>.+)"),
	OnGenExplainFn: func(cd *lib.ContextData) string {
		// TODO:
		if cd.Variables["reason"] == "'(' was never closed" {
			return "Your program did not close a parenthesis properly"
		}
		return "The interpreter was not able to understand your program because of an invalid syntax"
	},
	OnGenBugFixFn: func(cd *lib.ContextData) []lib.BugFix {
		// TODO
		return make([]lib.BugFix, 0)
	},
}
