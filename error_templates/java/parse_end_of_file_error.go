package java

import (
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

			s.AddStep("Add the missing `%s` in line %d", ctx.missingCharacter, endPos.Line+1).
				AddFix(lib.FixSuggestion{
					NewText:       "\n" + ctx.missingCharacter,
					StartPosition: endPos,
					EndPosition:   endPos,
				})
		})
	},
}

func nearestMissingNodeFromPos(cursor *sitter.TreeCursor, pos lib.Position) *sitter.Node {
	defer cursor.GoToParent()

	// hope it executes to avoid stack overflow
	if !cursor.GoToFirstChild() {
		return nil
	}

	for {
		currentNode := cursor.CurrentNode()
		pointA := currentNode.StartPoint()
		pointB := currentNode.EndPoint()

		if uint32(pos.Line) >= pointA.Row+1 && uint32(pos.Line) <= pointB.Row+1 {
			if currentNode.IsMissing() {
				return currentNode
			} else if currentNode.ChildCount() != 0 {
				if gotNode := nearestMissingNodeFromPos(cursor, pos); gotNode != nil {
					return gotNode
				}
			}
		}

		if !cursor.GoToNextSibling() {
			return nil
		}
	}
}
