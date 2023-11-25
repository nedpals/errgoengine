package java

import (
	"fmt"
	"strings"

	lib "github.com/nedpals/errgoengine"
)

type unclosedCharacterLiteralErrorCtx struct {
	parent lib.SyntaxNode
}

var UnclosedCharacterLiteralError = lib.ErrorTemplate{
	Name:              "UnclosedCharacterLiteralError",
	Pattern:           comptimeErrorPattern(`unclosed character literal`),
	StackTracePattern: comptimeStackTracePattern,
	OnAnalyzeErrorFn: func(cd *lib.ContextData, err *lib.MainError) {
		err.Context = unclosedCharacterLiteralErrorCtx{
			parent: err.Nearest,
		}

		if err.Nearest.Type() == "character_literal" {
			return
		}

		lib.QueryNode(err.Nearest, strings.NewReader("(character_literal) @literal"), func(ctx lib.QueryNodeCtx) bool {
			for _, c := range ctx.Match.Captures {
				node := lib.WrapNode(err.Document, c.Node)
				err.Nearest = node
				return false
			}
			return true
		})
	},
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		gen.Add("This error occurs when there's an attempt to define a character literal with more than one character, or if the character literal is not closed properly.")
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		ctx := cd.MainError.Context.(unclosedCharacterLiteralErrorCtx)
		valueNode := cd.MainError.Nearest

		if ctx.parent.Type() == "local_variable_declaration" {
			if isString := len(valueNode.Text()) > 1; isString {
				valueStartPos := valueNode.StartPosition()
				valueEndPos := valueNode.EndPosition()

				gen.Add("Store as a String", func(s *lib.BugFixSuggestion) {
					typeNode := ctx.parent.ChildByFieldName("type")

					s.AddStep("The character literal should contain only one character. If you intend to store a string, use double quotes (`\"`).").
						AddFix(lib.FixSuggestion{
							NewText:       "String",
							StartPosition: typeNode.StartPosition(),
							EndPosition:   typeNode.EndPosition(),
						}).
						AddFix(lib.FixSuggestion{
							NewText: "\"",
							StartPosition: lib.Position{
								Line:   valueStartPos.Line,
								Column: valueStartPos.Column,
							},
							EndPosition: lib.Position{
								Line:   valueEndPos.Line,
								Column: valueStartPos.Column + 1,
							},
						}).
						AddFix(lib.FixSuggestion{
							NewText: "\"",
							StartPosition: lib.Position{
								Line:   valueStartPos.Line,
								Column: valueEndPos.Column - 1,
							},
							EndPosition: lib.Position{
								Line:   valueEndPos.Line,
								Column: valueEndPos.Column,
							},
						})
				})

				gen.Add("Use single quotes for characters", func(s *lib.BugFixSuggestion) {
					s.AddStep("If you want to store a single character, ensure that you use single quotes (`'`).").
						AddFix(lib.FixSuggestion{
							NewText:       fmt.Sprintf("'%c'", valueNode.Text()[1]),
							StartPosition: valueStartPos,
							EndPosition:   valueEndPos,
						})
				})
			}
		}
	},
}
