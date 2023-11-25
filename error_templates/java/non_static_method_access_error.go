package java

import (
	"fmt"
	"strings"

	lib "github.com/nedpals/errgoengine"
)

type nonStaticMethodAccessErrorCtx struct {
	class  string
	method string
	parent lib.SyntaxNode
}

var NonStaticMethodAccessError = lib.ErrorTemplate{
	Name:              "NonStaticMethodAccessError",
	Pattern:           comptimeErrorPattern(`non-static method (?P<method>\S+)\(\) cannot be referenced from a static context`),
	StackTracePattern: comptimeStackTracePattern,
	OnAnalyzeErrorFn: func(cd *lib.ContextData, m *lib.MainError) {
		nCtx := nonStaticMethodAccessErrorCtx{parent: m.Nearest}

		// Get the class name
		symbols := cd.Symbols[cd.MainDocumentPath()]
		for _, sym := range symbols.Symbols {
			if sym.Kind() == lib.SymbolKindClass && m.Nearest.Location().IsWithin(sym.Location()) {
				nCtx.class = sym.Name()
				break
			}
		}

		m.Context = nCtx

		lib.QueryNode(m.Nearest, strings.NewReader("(method_invocation name: (identifier) @method arguments: (argument_list))"), func(ctx lib.QueryNodeCtx) bool {
			match := ctx.Cursor.FilterPredicates(ctx.Match, []byte(m.Nearest.Doc.Contents))
			for _, c := range match.Captures {
				node := lib.WrapNode(m.Nearest.Doc, c.Node)
				m.Nearest = node
				nCtx.method = node.Text()
				return false
			}
			return true
		})
	},
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		gen.Add("This error occurs when trying to access a non-static method from a static context. In Java, a non-static method belongs to an instance of the class and needs an object to be called upon.")
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		ctx := cd.MainError.Context.(nonStaticMethodAccessErrorCtx)

		gen.Add("Instantiate and call the method", func(s *lib.BugFixSuggestion) {
			s.AddStep("Create an instance of the class to access the non-static method").
				AddFix(lib.FixSuggestion{
					NewText: fmt.Sprintf("%s obj = new %s();\n", ctx.class, ctx.class),
					StartPosition: lib.Position{
						Line:   ctx.parent.StartPosition().Line,
						Column: ctx.parent.StartPosition().Column,
					},
					EndPosition: lib.Position{
						Line:   ctx.parent.EndPosition().Line,
						Column: 0,
					},
				}).
				AddFix(lib.FixSuggestion{
					NewText:       "obj.",
					StartPosition: ctx.parent.StartPosition(),
					EndPosition:   ctx.parent.StartPosition(),
				})
		})
	},
}
