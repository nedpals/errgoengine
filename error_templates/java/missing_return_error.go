package java

import (
	"fmt"
	"strings"

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
		rootNode := lib.WrapNode(m.Document, m.Document.Tree.RootNode())
		pos := m.ErrorNode.StartPos
		lib.QueryNode(rootNode, strings.NewReader("(method_declaration) @method"), func(ctx lib.QueryNodeCtx) bool {
			match := ctx.Cursor.FilterPredicates(ctx.Match, []byte(m.Nearest.Doc.Contents))
			for _, c := range match.Captures {
				pointA := c.Node.StartPoint()
				pointB := c.Node.EndPoint()
				if uint32(pos.Line) >= pointA.Row+1 && uint32(pos.Line) <= pointB.Row+1 {
					node := lib.WrapNode(m.Nearest.Doc, c.Node)
					mCtx.NearestMethod = node
					return false
				}
			}
			return true
		})
		fmt.Println(mCtx.NearestMethod.Text())
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

			s.AddStep(
				"Since the `%s` method is declared to return an `%s`, you need to provide a return statement with the result",
				ctx.NearestMethod.ChildByFieldName("name").Text(),
				ctx.NearestMethod.ChildByFieldName("type").Text(),
			).AddFix(lib.FixSuggestion{
				NewText:       "\n" + cd.MainError.Document.LineAt(lastStartPosInBlock.Line)[:lastStartPosInBlock.Column] + fmt.Sprintf("return %s;", ctx.NearestMethod.ChildByFieldName("type").Text()),
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
