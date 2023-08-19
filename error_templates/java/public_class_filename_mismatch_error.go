package java

import "github.com/nedpals/errgoengine/lib"

var PublicClassFilenameMismatchError = lib.ErrorTemplate{
	Name:              "PublicClassFilenameMismatchError",
	Pattern:           `(?P<stacktrace>(?:.|\s)*) error: class (?P<className>\S+) is public, should be declared in a file named (?P<classFileName>\S+\.java).*`,
	StackTracePattern: `(?P<path>\S+):(?P<position>\d+)`,
	OnGenExplainFn: func(cd *lib.ContextData) string {
		// TODO:
		return "TODo:"
	},
	OnGenBugFixFn: func(cd *lib.ContextData) []lib.BugFix {
		return make([]lib.BugFix, 0)
	},
}
