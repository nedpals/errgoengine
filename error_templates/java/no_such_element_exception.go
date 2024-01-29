package java

import (
	"fmt"

	lib "github.com/nedpals/errgoengine"
)

type noSuchElementExceptionCtx struct {
	parentNode      lib.SyntaxNode
	grandParentNode lib.SyntaxNode
}

var NoSuchElementException = lib.ErrorTemplate{
	Name:    "NoSuchElementException",
	Pattern: runtimeErrorPattern("java.util.NoSuchElementException", ""),
	OnAnalyzeErrorFn: func(cd *lib.ContextData, m *lib.MainError) {
		ctx := noSuchElementExceptionCtx{}
		query := `(method_invocation object: (_) name: (identifier) @fn-name arguments: (argument_list) (#eq? @fn-name "next")) @call`
		for q := m.Nearest.Query(query); q.Next(); {
			if q.CurrentTagName() != "fn-name" {
				continue
			}

			node := q.CurrentNode()
			// nCtx.ArrayExprNode = node
			// m.Nearest = node.ChildByFieldName("dimensions").NamedChild(0)
			m.Nearest = node
		}

		ctx.parentNode = m.Nearest.Parent() // expr -> method_invocation

		// get grandparent node
		ctx.grandParentNode = ctx.parentNode.Parent()
		if !ctx.grandParentNode.IsNull() {
			if ctx.grandParentNode.Type() == "expression_statement" || ctx.grandParentNode.Type() == "variable_declarator" {
				ctx.grandParentNode = ctx.grandParentNode.Parent()
			}
		}

		m.Context = ctx
	},
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		gen.Add("This error occurs when attempting to retrieve an element from an empty collection using an iterator.")
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		ctx := cd.MainError.Context.(noSuchElementExceptionCtx)
		// TODO: detect the statements that are using the variable to expand the position range of the if statement

		gen.Add("Check if the iterator has next elements before calling `next()`", func(s *lib.BugFixSuggestion) {
			step := s.AddStep("Ensure that the iterator has elements before attempting to retrieve the next one.")
			gpLocation := ctx.grandParentNode.Location()
			parentName := ctx.parentNode.ChildByFieldName("object").Text()
			// TODO: detect the statements that are using the variable to expand the position range of the if statement
			wrapStatement(
				step,
				fmt.Sprintf("if (%s.hasNext()) {", parentName),
				"\t} else {\n\t<i>System.out.println(\"No elements in the list.\");\n\t}",
				lib.Location{
					StartPos: lib.Position{
						Line: gpLocation.StartPos.Line,
					},
					EndPos: gpLocation.EndPos,
				},
				true,
			)
		})
	},
}
