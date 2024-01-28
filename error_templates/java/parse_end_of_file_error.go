package java

import (
	"strings"

	lib "github.com/nedpals/errgoengine"
	sitter "github.com/smacker/go-tree-sitter"
)

type parseEofErrorCtx struct {
	missingSymStack  []string
	missingCharacter string
}

var ParseEndOfFileError = lib.ErrorTemplate{
	Name:              "ParseEndOfFileError",
	Pattern:           comptimeErrorPattern("reached end of file while parsing"),
	StackTracePattern: comptimeStackTracePattern,
	OnAnalyzeErrorFn: func(cd *lib.ContextData, m *lib.MainError) {
		ctx := parseEofErrorCtx{}

		// traverse the tree and get the nearest "missing" node
		rootNode := m.Document.Tree.RootNode()
		cursor := sitter.NewTreeCursor(rootNode)
		rawNearestMissingNode := nearestMissingNodeFromPos(cursor, m.ErrorNode.StartPos)
		nearestMissingNode := lib.WrapNode(m.Document, rawNearestMissingNode)
		m.Nearest = nearestMissingNode
		nearestStr := m.Nearest.String()
		prefix := "(MISSING \""
		ctx.missingCharacter = nearestStr[len(prefix) : len(prefix)+1]
		m.Context = ctx
	},
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		gen.Add("This error occurs when the compiler expects more code but encounters the end of the file.")
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		ctx := cd.MainError.Context.(parseEofErrorCtx)

		gen.Add("Complete the code", func(s *lib.BugFixSuggestion) {
			endPos := cd.MainError.Nearest.EndPosition()
			endLine := endPos.Line
			space := ""

			// add the missing at the right depth so we need to identify first
			// where should we add the missing character
			for line := endPos.Line; line >= 0; line-- {
				if strings.TrimSpace(cd.MainError.Document.LineAt(line)) != ctx.missingCharacter {
					endLine = line
					break
				}

				if len(space) == 0 {
					space = getSpaceFromBeginning(cd.MainError.Document, line, endPos.Column)
				}

				continue
			}

			s.AddStep("Add the missing `%s` in line %d", ctx.missingCharacter, endPos.Line+1).
				AddFix(lib.FixSuggestion{
					NewText: "\n" + getIndent(space, endPos.Line-endLine) + ctx.missingCharacter,
					StartPosition: lib.Position{
						Line: endLine,
					},
					EndPosition: lib.Position{
						Line: endLine,
					},
				})
		})
	},
}
