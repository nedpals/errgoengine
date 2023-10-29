package python

import lib "github.com/nedpals/errgoengine"

var ValueError = lib.ErrorTemplate{
	Name:    "ValueError",
	Pattern: "ValueError: (?P<reason>.+)",
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		// TODO:
		if cd.Variables["reason"] == "invalid literal for int() with base 10: 'abc'" {
			gen.Add("The input provided is an alphabetical string which cannot be converted into an int")
		}
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		// TODO
	},
}
