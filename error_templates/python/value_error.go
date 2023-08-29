package python

import lib "github.com/nedpals/errgoengine"

var ValueError = lib.ErrorTemplate{
	Name:    "ValueError",
	Pattern: "ValueError: (?P<reason>.+)",
	OnGenExplainFn: func(cd *lib.ContextData) string {
		// TODO:
		if cd.Variables["reason"] == "invalid literal for int() with base 10: 'abc'" {
			return "The input provided is an alphabetical string which cannot be converted into an int"
		}
		return "Unknown valueerror"
	},
	OnGenBugFixFn: func(cd *lib.ContextData) []lib.BugFix {
		// TODO
		return make([]lib.BugFix, 0)
	},
}
