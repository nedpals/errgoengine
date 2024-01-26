package java

import (
	"fmt"
	"strconv"

	lib "github.com/nedpals/errgoengine"
)

type stringIndexOutOfBoundsExceptionCtx struct {
	parentNode      lib.SyntaxNode
	grandParentNode lib.SyntaxNode
}

var StringIndexOutOfBoundsException = lib.ErrorTemplate{
	Name:    "StringIndexOutOfBoundsException",
	Pattern: runtimeErrorPattern("java.lang.StringIndexOutOfBoundsException", `String index out of range: (?P<index>\d+)`),
	OnAnalyzeErrorFn: func(cd *lib.ContextData, m *lib.MainError) {
		ctx := stringIndexOutOfBoundsExceptionCtx{}

		for q := m.Nearest.Query(`(method_invocation name: (identifier) @name arguments: (argument_list (_) @arg) (#eq? @name "charAt") (#any-eq? @arg "%s"))`, cd.Variables["index"]); q.Next(); {
			if q.CurrentTagName() != "arg" {
				continue
			}
			node := q.CurrentNode()
			m.Nearest = node
			ctx.parentNode = node.Parent().Parent() // expr -> argument_list -> method_invocation

			// get grandparent node
			ctx.grandParentNode = ctx.parentNode.Parent()
			if !ctx.grandParentNode.IsNull() {
				if ctx.grandParentNode.Type() == "expression_statement" || ctx.grandParentNode.Type() == "variable_declarator" {
					ctx.grandParentNode = ctx.grandParentNode.Parent()
				}
			}
			break
		}

		m.Context = ctx
	},
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		gen.Add("This error occurs because the code is trying to access index %s that is beyond the length of the string.", cd.Variables["index"])
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		ctx := cd.MainError.Context.(stringIndexOutOfBoundsExceptionCtx)
		// arrayLen, _ := strconv.Atoi(cd.Variables["length"])
		index, _ := strconv.Atoi(cd.Variables["index"])
		// symbolTree := cd.Store.InitOrGetSymbolTree(cd.MainError.DocumentPath())

		gen.Add("Ensure the index is within the string length", func(s *lib.BugFixSuggestion) {
			obj := ctx.parentNode.ChildByFieldName("object")
			step := s.AddStep("Check that the index used for accessing the character is within the valid range of the string length.")
			gpLocation := ctx.grandParentNode.Location()

			// TODO: detect the statements that are using the variable to expand the position range of the if statement
			wrapWithCondStatement(
				step,
				cd.MainError.Document,
				"if",
				fmt.Sprintf("%d < %s.length()", index, obj.Text()),
				gpLocation,
				false,
			)

			wrapWithCondStatement(
				step,
				cd.MainError.Document,
				"else",
				"",
				lib.Location{
					StartPos: gpLocation.EndPos,
					EndPos:   gpLocation.EndPos,
				},
				true,
			)

			space := getSpace(
				cd.MainError.Document,
				gpLocation.StartPos.Line, 0, gpLocation.StartPos.Column, true)

			step.AddFix(lib.FixSuggestion{
				NewText: indentSpace(space, 1) + `System.out.println("Index out of range.")`,
				StartPosition: lib.Position{
					Line: gpLocation.EndPos.Line - 2,
				},
				EndPosition: lib.Position{
					Line: gpLocation.EndPos.Line - 2,
				},
			})
		})
	},
}
