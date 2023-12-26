package python

import (
	"fmt"
	"strings"

	lib "github.com/nedpals/errgoengine"
)

var NameError = lib.ErrorTemplate{
	Name:    "NameError",
	Pattern: `NameError: name '(?P<variable>\S+)' is not defined`,
	OnAnalyzeErrorFn: func(cd *lib.ContextData, m *lib.MainError) {
		lib.QueryNode(m.Nearest, strings.NewReader(fmt.Sprintf(`((identifier) @name (#eq? @name "%s"))`, cd.Variables["variable"])), func(ctx lib.QueryNodeCtx) bool {
			match := ctx.Cursor.FilterPredicates(ctx.Match, []byte(m.Nearest.Doc.Contents))
			for _, c := range match.Captures {
				node := lib.WrapNode(m.Nearest.Doc, c.Node)
				m.Nearest = node
				return false
			}
			return true
		})
	},
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		gen.Add("This error occurs when trying to use a variable (`%s`) or name that has not been defined in the current scope.", cd.Variables["variable"])
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		gen.Add("Define the variable before using it", func(s *lib.BugFixSuggestion) {
			// get to the very parent (before `block`)
			parent := cd.MainError.Nearest.Parent()
			for !parent.IsNull() && parent.Type() != "module" && parent.Type() != "block" {
				fmt.Println(parent.Type())
				parent = parent.Parent()
			}

			spaces := cd.MainError.Document.LineAt(parent.StartPosition().Line)[:parent.StartPosition().Column]

			s.AddStep("Make sure to define the variable `%s` before using it.", cd.Variables["variable"]).
				AddFix(lib.FixSuggestion{
					NewText: spaces + fmt.Sprintf("%s = \"Hello!\"\n", cd.Variables["variable"]),
					StartPosition: lib.Position{
						Line: parent.StartPosition().Line,
					},
				})
		})
	},
}
