package python

import lib "github.com/nedpals/errgoengine"

var IndentationError = lib.ErrorTemplate{
	Name:    "IndentationError",
	Pattern: compileTimeError("IndentationError: unindent does not match any outer indentation level"),
	OnGenExplainFn: func(cd *lib.ContextData) string {
		return "The code is not indented properly"
	},
	OnGenBugFixFn: func(cd *lib.ContextData) []lib.BugFix {
		// TODO:
		return make([]lib.BugFix, 0)
	},
}
