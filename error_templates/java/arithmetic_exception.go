package java

import lib "github.com/nedpals/errgoengine"

var ArithmeticException = lib.ErrorTemplate{
	Name:    "ArithmeticException",
	Pattern: runtimeErrorPattern("java.lang.ArithmeticException", "(?P<reason>.+)"),
	OnGenExplainFn: func(cd *lib.ContextData) string {
		reason := cd.Variables["reason"]
		switch reason {
		case "/ by zero":
			return "One of your variables initialized a double value by dividing a number to zero"
		case "Non-terminating decimal expansion; no exact representable decimal result.":
			return "TODO"
		default:
			return "Unknown ArithmeticException"
		}
	},
	OnGenBugFixFn: func(cd *lib.ContextData) []lib.BugFix {
		// TODO:
		return make([]lib.BugFix, 0)
	},
}
