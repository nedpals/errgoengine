package java

import (
	"fmt"
	"strings"

	lib "github.com/nedpals/errgoengine"
)

type unknownVarErrorCtx struct {
	rootNode   lib.SyntaxNode
	parentNode lib.SyntaxNode
}

var UnknownVariableError = lib.ErrorTemplate{
	Name:              "UnknownVariableError",
	Pattern:           comptimeErrorPattern("cannot find symbol", `symbol:\s+variable (?P<variable>\S+)`),
	StackTracePattern: comptimeStackTracePattern,
	OnAnalyzeErrorFn: func(cd *lib.ContextData, m *lib.MainError) {
		variable := cd.Variables["variable"]
		query := fmt.Sprintf("((identifier) @symbol (#eq? @symbol \"%s\"))", variable)

		lib.QueryNode(m.Nearest, strings.NewReader(query), func(ctx lib.QueryNodeCtx) bool {
			match := ctx.Cursor.FilterPredicates(ctx.Match, []byte(m.Nearest.Doc.Contents))
			for _, c := range match.Captures {
				node := lib.WrapNode(m.Nearest.Doc, c.Node)
				m.Context = unknownVarErrorCtx{m.Nearest, node.Parent()}
				m.Nearest = node
				return false
			}
			return true
		})
	},
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		gen.Add(`The program cannot find variable "%s"`, cd.Variables["variable"])
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		ctx := cd.MainError.Context.(unknownVarErrorCtx)
		variable := cd.Variables["variable"]

		gen.Add("Create a variable.", func(s *lib.BugFixSuggestion) {
			s.AddStep("Create a variable named \"%s\". For example:", variable).
				// TODO: use variable type from the inferred parameter
				AddFix(lib.FixSuggestion{
					NewText:       fmt.Sprintf("String %s = \"\";\n", variable),
					StartPosition: ctx.rootNode.StartPosition(),
					EndPosition:   ctx.rootNode.StartPosition(),
				})
		})
	},
}
