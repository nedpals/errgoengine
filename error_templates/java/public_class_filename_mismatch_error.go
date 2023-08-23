package java

import lib "github.com/nedpals/errgoengine"

var PublicClassFilenameMismatchError = lib.ErrorTemplate{
	Name:              "PublicClassFilenameMismatchError",
	Pattern:           comptimeErrorPattern(`class (?P<className>\S+) is public, should be declared in a file named (?P<classFileName>\S+\.java)`),
	StackTracePattern: comptimeStackTracePattern,
	OnGenExplainFn: func(cd *lib.ContextData) string {
		// TODO:
		return "TODo:"
	},
	OnGenBugFixFn: func(cd *lib.ContextData) []lib.BugFix {
		return make([]lib.BugFix, 0)
	},
}
