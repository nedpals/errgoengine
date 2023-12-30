package java

import (
	"fmt"

	lib "github.com/nedpals/errgoengine"
)

type invalidMethodDeclarationErrorCtx struct {
	declNode        lib.SyntaxNode
	returnTypeToAdd lib.Symbol
}

var InvalidMethodDeclarationError = lib.ErrorTemplate{
	Name:              "InvalidMethodDeclarationError",
	Pattern:           comptimeErrorPattern("invalid method declaration; return type required"),
	StackTracePattern: comptimeStackTracePattern,
	OnAnalyzeErrorFn: func(cd *lib.ContextData, m *lib.MainError) {
		iCtx := invalidMethodDeclarationErrorCtx{}
		pos := m.ErrorNode.StartPos

		for q := m.Document.RootNode().Query("(constructor_declaration) @method"); q.Next(); {
			node := q.CurrentNode()
			pointA := node.StartPoint()
			pointB := node.EndPoint()
			if uint32(pos.Line) >= pointA.Row+1 && uint32(pos.Line) <= pointB.Row+1 {
				iCtx.declNode = node
				m.Nearest = node.ChildByFieldName("name")
				break
			}
		}

		iCtx.returnTypeToAdd = lib.UnwrapReturnType(cd.FindSymbol(m.Nearest.Text(), m.Nearest.StartPosition().Index))
		m.Context = iCtx
	},
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		gen.Add("This error occurs when there is an invalid method declaration, and a return type is missing.")
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		ctx := cd.MainError.Context.(invalidMethodDeclarationErrorCtx)

		gen.Add("Add the return type to the method declaration", func(s *lib.BugFixSuggestion) {
			s.AddStep("Specify the return type of the `%s` method", cd.MainError.Nearest.Text()).
				AddFix(lib.FixSuggestion{
					NewText:       fmt.Sprintf("%s ", ctx.returnTypeToAdd.Name()),
					StartPosition: cd.MainError.Nearest.StartPosition(),
					EndPosition:   cd.MainError.Nearest.StartPosition(),
				})
		})
	},
}
