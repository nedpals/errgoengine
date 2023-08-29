package java

import lib "github.com/nedpals/errgoengine"

var ArithmeticException = lib.ErrorTemplate{
	Name:    "ArithmeticException",
	Pattern: runtimeErrorPattern("java.lang.ArithmeticException", "(?P<reason>.+)"),
	OnGenExplainFn: func(cd *lib.ContextData) string {
		if cd.Variables["reason"] == "/ by zero" {
			// TODO:
			return "One of your variables initialized a double value by dividing a number to zero"
		}

		// TODO:
		return "arithmeticexception todo!"
	},
	OnGenBugFixFn: func(cd *lib.ContextData) []lib.BugFix {
		// TODO:
		return make([]lib.BugFix, 0)
	},
}
