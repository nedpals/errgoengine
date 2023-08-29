package python

import lib "github.com/nedpals/errgoengine"

var ZeroDivisionError = lib.ErrorTemplate{
	Name:    "ZeroDivisionError",
	Pattern: "ZeroDivisionError: division by zero",
	OnGenExplainFn: func(cd *lib.ContextData) string {
		// TODO:
		return "The number has been divided by zero"
	},
	OnGenBugFixFn: func(cd *lib.ContextData) []lib.BugFix {
		// TODO:
		return make([]lib.BugFix, 0)
	},
}
