package python

import (
	"strings"

	lib "github.com/nedpals/errgoengine"
)

type valueErrorKind int

const (
	valueErrorKindUnknown valueErrorKind = 0
	valueErrorKindInt     valueErrorKind = iota
)

type valueErrorCtx struct {
	kind     valueErrorKind
	callNode lib.SyntaxNode
}

var ValueError = lib.ErrorTemplate{
	Name:    "ValueError",
	Pattern: "ValueError: (?P<reason>.+)",
	OnAnalyzeErrorFn: func(cd *lib.ContextData, m *lib.MainError) {
		vCtx := valueErrorCtx{}
		reason := cd.Variables["reason"]
		query := ""

		if strings.HasPrefix(reason, "invalid literal for int() with base 10") {
			vCtx.kind = valueErrorKindInt
			query = "((call function: (identifier) @func) @call (#eq? @func \"int\"))"
		} else {
			vCtx.kind = valueErrorKindUnknown
		}

		if len(query) != 0 {
			lib.QueryNode(m.Nearest, strings.NewReader(query), func(ctx lib.QueryNodeCtx) bool {
				match := ctx.Cursor.FilterPredicates(ctx.Match, []byte(m.Nearest.Doc.Contents))
				for _, c := range match.Captures {
					node := lib.WrapNode(m.Nearest.Doc, c.Node)
					vCtx.callNode = node
					m.Nearest = node.ChildByFieldName("arguments").FirstNamedChild()
					return false
				}
				return true
			})
		}

		m.Context = vCtx
	},
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		ctx := cd.MainError.Context.(valueErrorCtx)

		switch ctx.kind {
		case valueErrorKindInt:
			gen.Add("This error occurs when you try to convert a value to `int`, but the value is not a valid `int`.")
		default:
			gen.Add("This error occurs when you try to convert a value to another type, but the value is not a valid value for that type.")
		}
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		ctx := cd.MainError.Context.(valueErrorCtx)

		switch ctx.kind {
		case valueErrorKindInt:
			gen.Add("Use a valid integer string", func(s *lib.BugFixSuggestion) {
				s.AddStep("Make sure the value you're trying to convert is a valid integer string.").
					AddFix(lib.FixSuggestion{
						NewText:       `"123"`,
						StartPosition: cd.MainError.Nearest.StartPosition(),
						EndPosition:   cd.MainError.Nearest.EndPosition(),
					})
			})
		}

		gen.Add("Add error handling", func(s *lib.BugFixSuggestion) {
			parent := ctx.callNode.Parent()
			for !parent.IsNull() && parent.Type() != "block" && parent.Type() != "module" {
				parent = parent.Parent()
			}

			spaces := cd.MainError.Document.LineAt(parent.StartPosition().Line)[:parent.StartPosition().Column]
			origSpaces := strings.Clone(spaces)

			if len(spaces) == 0 {
				spaces = "\t"
			}

			s.AddStep("To handle invalid inputs gracefully, you can use a try-except block.").
				AddFix(lib.FixSuggestion{
					NewText: origSpaces + "try:\n",
					StartPosition: lib.Position{
						Line: parent.StartPosition().Line,
					},
					EndPosition: lib.Position{
						Line: parent.StartPosition().Line,
					},
				}).
				AddFix(lib.FixSuggestion{
					NewText:       spaces,
					StartPosition: parent.StartPosition(),
					EndPosition:   parent.StartPosition(),
				}).
				AddFix(lib.FixSuggestion{
					NewText:       origSpaces + "except ValueError as e:\n" + spaces + "print(f\"Error: {e}\")",
					StartPosition: parent.EndPosition(),
					EndPosition:   parent.EndPosition(),
				})
		})
	},
}
