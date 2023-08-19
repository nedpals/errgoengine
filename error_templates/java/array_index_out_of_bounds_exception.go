package java

import (
	"fmt"

	"github.com/nedpals/errgoengine/lib"
)

var ArrayIndexOutOfBoundsException = lib.ErrorTemplate{
	Name:    "ArrayIndexOutOfBoundsException",
	Pattern: `Exception in thread "(?P<thread>\w+)" java\.lang\.ArrayIndexOutOfBoundsException: Index (?P<index>\d+) out of bounds for length (?P<length>\d+)`,
	OnGenExplainFn: func(cd *lib.ContextData) string {
		// TODO:

		return fmt.Sprintf("Your program attempted to access an element in index %s on an array that has only %s items", cd.Variables["index"], cd.Variables["length"])
	},
	OnGenBugFixFn: func(cd *lib.ContextData) []lib.BugFix {
		return []lib.BugFix{}
	},
}
