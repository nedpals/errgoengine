package java

import (
	"path/filepath"
	"strings"

	lib "github.com/nedpals/errgoengine"
)

type publicClassFilenameMismatchErrorCtx struct {
	className             string
	expectedClassName     string
	actualClassFileName   string
	expectedClassFilename string
}

var PublicClassFilenameMismatchError = lib.ErrorTemplate{
	Name:              "PublicClassFilenameMismatchError",
	Pattern:           comptimeErrorPattern(`class (?P<className>\S+) is public, should be declared in a file named (?P<classFileName>\S+\.java)`),
	StackTracePattern: comptimeStackTracePattern,
	OnAnalyzeErrorFn: func(cd *lib.ContextData, m *lib.MainError) {
		className := cd.Variables["className"]
		actualClassFilename := filepath.Base(cd.MainError.ErrorNode.DocumentPath)
		m.Context = publicClassFilenameMismatchErrorCtx{
			className:             className,
			actualClassFileName:   actualClassFilename,                                  // the name from the filename
			expectedClassName:     strings.Replace(actualClassFilename, ".java", "", 1), // the name from the filename without the extension
			expectedClassFilename: className + ".java",                                  // the expected filename to be renamed
		}

		for q := m.Nearest.Doc.RootNode().Query(`(class_declaration name: (identifier) @class-name (#eq? @class-name "%s"))`, className); q.Next(); {
			node := q.CurrentNode()
			m.Nearest = node
			break
		}
	},
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		gen.Add(`This error occurs because the name of the Java file does not match the name of the public class within it.`)
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		ctx := cd.MainError.Context.(publicClassFilenameMismatchErrorCtx)

		gen.Add("Rename your file", func(s *lib.BugFixSuggestion) {
			s.AddStep("Rename the file to \"%s\" to match the class", ctx.expectedClassFilename)
		})

		gen.Add("Rename the public class", func(s *lib.BugFixSuggestion) {
			s.AddStep("The filename should match the name of the public class in the file. To resolve this, change the class name to match the filename.").
				AddFix(lib.FixSuggestion{
					NewText:       ctx.expectedClassName,
					StartPosition: cd.MainError.Nearest.StartPosition(),
					EndPosition:   cd.MainError.Nearest.EndPosition(),
				})
		})
	},
}
