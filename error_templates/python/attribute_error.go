package python

import (
	"fmt"

	lib "github.com/nedpals/errgoengine"
)

var AttributeError = lib.ErrorTemplate{
	Name:    "AttributeError",
	Pattern: `AttributeError: '(?P<typeName>\S+)' object has no attribute '(?P<method>\S+)'`,
	OnGenExplainFn: func(cd *lib.ContextData) string {
		// TODO:
		return fmt.Sprintf(`Method "%s" does not exist in "%s" type`, cd.Variables["method"], cd.Variables["typeName"])
	},
	OnGenBugFixFn: func(cd *lib.ContextData) []lib.BugFix {
		// TODO:
		return make([]lib.BugFix, 0)
	},
}
