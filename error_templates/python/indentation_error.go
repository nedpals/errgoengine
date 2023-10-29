package python

import lib "github.com/nedpals/errgoengine"

var IndentationError = lib.ErrorTemplate{
	Name:    "IndentationError",
	Pattern: compileTimeError("IndentationError: unindent does not match any outer indentation level"),
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		gen.Add("The code is not indented properly")
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		// TODO:
	},
}
