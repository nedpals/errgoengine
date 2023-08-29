package java

import (
	"fmt"

	lib "github.com/nedpals/errgoengine"
)

var UnknownVariableError = lib.ErrorTemplate{
	Name:              "UnknownVariableError",
	Pattern:           comptimeErrorPattern("cannot find symbol", `symbol:\s+variable (?P<variable>\S+)`),
	StackTracePattern: comptimeStackTracePattern,
	OnGenExplainFn: func(cd *lib.ContextData) string {
		return fmt.Sprintf(`The program cannot find variable "%s"`, cd.Variables["variable"])
	},
	OnGenBugFixFn: func(cd *lib.ContextData) []lib.BugFix {
		// TODO:
		return make([]lib.BugFix, 0)
	},
}
