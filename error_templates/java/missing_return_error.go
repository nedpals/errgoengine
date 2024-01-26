package java

import (
	"context"
	"fmt"

	lib "github.com/nedpals/errgoengine"
)

type missingReturnErrorCtx struct {
	NearestMethod lib.SyntaxNode
}

var MissingReturnError = lib.ErrorTemplate{
	Name:              "MissingReturnError",
	Pattern:           comptimeErrorPattern(`missing return statement`),
	StackTracePattern: comptimeStackTracePattern,
	OnAnalyzeErrorFn: func(cd *lib.ContextData, m *lib.MainError) {
		// get nearest method declaration
		mCtx := missingReturnErrorCtx{}
		rootNode := m.Document.RootNode()
		pos := m.ErrorNode.StartPos

		for q := rootNode.Query("(method_declaration) @method"); q.Next(); {
			node := q.CurrentNode()
			pointA := node.StartPoint()
			pointB := node.EndPoint()
			if uint32(pos.Line) >= pointA.Row+1 && uint32(pos.Line) <= pointB.Row+1 {
				mCtx.NearestMethod = node
				break
			}
		}

		m.Context = mCtx
	},
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		gen.Add("This error occurs when a method is declared to return a value, but there is no return statement within the method.")
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		ctx := cd.MainError.Context.(missingReturnErrorCtx)

		// TODO
		gen.Add("Provide a return statement", func(s *lib.BugFixSuggestion) {
			bodyNode := ctx.NearestMethod.ChildByFieldName("body")
			lastStartPosInBlock := bodyNode.EndPosition()
			lastEndPosInBlock := bodyNode.EndPosition()
			if bodyNode.NamedChildCount() > 0 {
				lastStartPosInBlock = bodyNode.LastNamedChild().StartPosition()
				lastEndPosInBlock = bodyNode.LastNamedChild().EndPosition()
			}

			expectedTypeNode := ctx.NearestMethod.ChildByFieldName("type")
			expectedTypeSym := cd.Analyzer.AnalyzeNode(context.Background(), expectedTypeNode)
			nearestScope := cd.InitOrGetSymbolTree(cd.MainDocumentPath()).GetNearestScopedTree(lastEndPosInBlock.Index)
			symbolsForReturn := nearestScope.FindSymbolsByClause(func(sym lib.Symbol) bool {
				if sym, ok := sym.(lib.IReturnableSymbol); ok {
					return sym.ReturnType() == expectedTypeSym
				}
				return false
			})

			// nearest sym will be at the last
			valueToReturn := getDefaultValueForType(expectedTypeSym)
			if len(symbolsForReturn) != 0 {
				nearestSym := symbolsForReturn[len(symbolsForReturn)-1]
				valueToReturn = nearestSym.Name()
			}

			s.AddStep(
				"Since the `%s` method is declared to return an `%s`, you need to provide a return statement with the result.",
				ctx.NearestMethod.ChildByFieldName("name").Text(),
				ctx.NearestMethod.ChildByFieldName("type").Text(),
			).AddFix(lib.FixSuggestion{
				NewText:       "\n" + getSpaceFromBeginning(cd.MainError.Document, lastStartPosInBlock.Line, lastStartPosInBlock.Column) + fmt.Sprintf("return %s;", valueToReturn),
				StartPosition: lastEndPosInBlock,
				EndPosition:   lastEndPosInBlock,
				Description:   "This ensures that the method returns the sum of the two input numbers.",
			})
		})

		gen.Add("Set the method return type to void", func(s *lib.BugFixSuggestion) {
			s.AddStep(
				"If you don't intend to return a value from the `%s` method, you can change its return type to `void`.",
				ctx.NearestMethod.ChildByFieldName("name").Text(),
			).AddFix(lib.FixSuggestion{
				NewText:       "void",
				StartPosition: ctx.NearestMethod.ChildByFieldName("type").StartPosition(),
				EndPosition:   ctx.NearestMethod.ChildByFieldName("type").EndPosition(),
				Description:   "This is appropriate if you're using the method for side effects rather than returning a value.",
			})
		})
	},
}
