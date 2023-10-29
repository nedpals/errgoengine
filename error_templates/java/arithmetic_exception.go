package java

import lib "github.com/nedpals/errgoengine"

var ArithmeticException = lib.ErrorTemplate{
	Name:    "ArithmeticException",
	Pattern: runtimeErrorPattern("java.lang.ArithmeticException", "(?P<reason>.+)"),
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		reason := cd.Variables["reason"]
		switch reason {
		case "/ by zero":
			gen.Add("One of your variables initialized a double value by dividing a number to zero")
		case "Non-terminating decimal expansion; no exact representable decimal result.":
			gen.Add("TODO")
		default:
			gen.Add("Unknown ArithmeticException")
		}
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		// TODO:
	},
}
