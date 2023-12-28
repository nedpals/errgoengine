package java

import (
	"fmt"
	"strings"

	lib "github.com/nedpals/errgoengine"
)

var ArrayRequiredTypeError = lib.ErrorTemplate{
	Name:              "ArrayRequiredTypeError",
	Pattern:           comptimeErrorPattern(`array required, but (?P<foundType>\S+) found`),
	StackTracePattern: comptimeStackTracePattern,
	OnAnalyzeErrorFn: func(cd *lib.ContextData, err *lib.MainError) {
		query := strings.NewReader("(array_access array: (identifier) index: ((_) @index (#eq? @index \"0\")))")
		lib.QueryNode(cd.MainError.Nearest, query, func(ctx lib.QueryNodeCtx) bool {
			match := ctx.Cursor.FilterPredicates(ctx.Match, []byte(cd.MainError.Nearest.Doc.Contents))
			for _, c := range match.Captures {
				node := lib.WrapNode(cd.MainError.Nearest.Doc, c.Node)
				err.Nearest = node
				return false
			}
			return true
		})
	},
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		parent := cd.MainError.Nearest.Parent()
		varNode := parent.ChildByFieldName("array")
		indexNode := cd.MainError.Nearest

		gen.Add(
			"This error occurs because the variable `%s` is declared as an `%s` rather than an array. You're attempting to access an index (`%s`) on a variable that's not an array.",
			varNode.Text(),
			cd.Variables["foundType"],
			indexNode.Text(),
		)
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		parent := cd.MainError.Nearest.Parent()
		varNode := parent.ChildByFieldName("array")
		// indexNode := cd.MainError.Nearest
		tree := cd.InitOrGetSymbolTree(cd.MainDocumentPath())

		gen.Add("Convert variable to an array", func(s *lib.BugFixSuggestion) {
			declSym := tree.GetSymbolByNode(getIdentifierNode(varNode))
			declNode := cd.MainError.Document.RootNode().NamedDescendantForPointRange(
				lib.Location{
					StartPos: declSym.Location().StartPos,
					EndPos:   declSym.Location().StartPos,
				},
			).Parent()

			valueNode := declNode.ChildByFieldName("value")
			declNode = declNode.Parent()

			s.AddStep("Declare the variable `%s` as an array of `%s`.", varNode.Text(), cd.Variables["foundType"]).
				AddFix(lib.FixSuggestion{
					NewText:       fmt.Sprintf("%s[] %s = {%s};", cd.Variables["foundType"], varNode.Text(), valueNode.Text()),
					StartPosition: declNode.StartPosition(),
					EndPosition:   declNode.EndPosition(),
				})
		})
	},
}
