package python

import lib "github.com/nedpals/errgoengine"

var ZeroDivisionError = lib.ErrorTemplate{
	Name:    "ZeroDivisionError",
	Pattern: "ZeroDivisionError: division by zero",
	OnGenExplainFn: func(cd *lib.ContextData) string {
		return "Your program did something nasty dawg"
	},
	OnGenBugFixFn: func(cd *lib.ContextData) []lib.BugFix {
		// TODO:
		return make([]lib.BugFix, 0)
	},
}
