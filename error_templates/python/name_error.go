package python

import (
	"fmt"

	"github.com/nedpals/errgoengine/lib"
)

var NameError = lib.ErrorTemplate{
	Name:    "NameError",
	Pattern: `NameError: name '(?P<variable>\S+)' is not defined`,
	OnGenExplainFn: func(cd *lib.ContextData) string {
		return fmt.Sprintf("Your program tried to access the '%s' variable which was not found on your program.", cd.Variables["variable"])
	},
	OnGenBugFixFn: func(cd *lib.ContextData) []lib.BugFix {
		// TODO:
		return make([]lib.BugFix, 0)
	},
}
