package java

import (
	"fmt"
	"strings"

	lib "github.com/nedpals/errgoengine"
)

type privateAccessErrorCtx struct {
	ClassDeclarationNode lib.SyntaxNode
}

var PrivateAccessError = lib.ErrorTemplate{
	Name:              "PrivateAccessError",
	Pattern:           comptimeErrorPattern(`(?P<field>\S+) has private access in (?P<class>\S+)`),
	StackTracePattern: comptimeStackTracePattern,
	OnAnalyzeErrorFn: func(cd *lib.ContextData, m *lib.MainError) {
		pCtx := privateAccessErrorCtx{}
		className := cd.Variables["class"]
		rootNode := lib.WrapNode(m.Nearest.Doc, m.Nearest.Doc.Tree.RootNode())

		// locate the right node first
		query := fmt.Sprintf(`((field_access (identifier) . (identifier) @field-name) @field (#eq? @field-name "%s"))`, cd.Variables["field"])
		lib.QueryNode(rootNode, strings.NewReader(query), func(ctx lib.QueryNodeCtx) bool {
			match := ctx.Cursor.FilterPredicates(ctx.Match, []byte(m.Nearest.Doc.Contents))
			for _, c := range match.Captures {
				node := lib.WrapNode(m.Nearest.Doc, c.Node)
				m.Nearest = node
				return false
			}
			return true
		})

		// get class declaration node
		classQuery := fmt.Sprintf(`(class_declaration name: (identifier) @class-name (#eq? @class-name "%s"))`, className)
		lib.QueryNode(rootNode, strings.NewReader(classQuery), func(ctx lib.QueryNodeCtx) bool {
			match := ctx.Cursor.FilterPredicates(ctx.Match, []byte(m.Nearest.Doc.Contents))
			for _, c := range match.Captures {
				node := lib.WrapNode(m.Nearest.Doc, c.Node)
				pCtx.ClassDeclarationNode = node
				return false
			}
			return true
		})

		m.Context = pCtx
	},
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		gen.Add("This error occurs when you try to access a private variable from another class, which is not allowed.")
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		// ctx := cd.MainError.Context.(privateAccessErrorCtx)
		fmt.Println(cd.MainError.Nearest.String())
		fmt.Println(cd.Analyzer.AnalyzeNode(cd.MainError.Nearest))

		gen.Add("Use a public accessor method", func(s *lib.BugFixSuggestion) {
			// methodCreatorSb := &strings.Builder{}

			// get return type of the private field

			// methodCreatorSb.WriteString("public ")

			// s.AddStep("To access a private variable from another class, create a public accessor method in `%s`", cd.Variables["class"]).
			// 	AddFix()
		})
	},
}
