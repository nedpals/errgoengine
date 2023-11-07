package java

import (
	"path/filepath"

	lib "github.com/nedpals/errgoengine"
)

var PublicClassFilenameMismatchError = lib.ErrorTemplate{
	Name:              "PublicClassFilenameMismatchError",
	Pattern:           comptimeErrorPattern(`class (?P<className>\S+) is public, should be declared in a file named (?P<classFileName>\S+\.java)`),
	StackTracePattern: comptimeStackTracePattern,
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		gen.Add(
			`Public class "%s" does not match the name of the file which is "%s". Rename it to "%s"`,
			cd.Variables["className"],
			filepath.Base(cd.MainError.ErrorNode.DocumentPath),
			cd.Variables["classFileName"],
		)
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {

	},
}
