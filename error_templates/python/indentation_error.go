package python

import (
	"fmt"

	lib "github.com/nedpals/errgoengine"
)

var IndentationError = lib.ErrorTemplate{
	Name:    "IndentationError",
	Pattern: compileTimeError("IndentationError: unindent does not match any outer indentation level"),
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		gen.Add("This error occurs when there is a mismatch in the indentation levels in the code.")
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		gen.Add("Correct the indentation", func(s *lib.BugFixSuggestion) {
			// use the previous sibling's spacing as basis
			prevSibling := cd.MainError.Nearest.PrevNamedSibling()
			if prevSibling.Type() == "function_definition" {
				prevSibling = prevSibling.ChildByFieldName("body").LastNamedChild()
			}

			fmt.Println(prevSibling.String())
			spaces := cd.MainError.Document.LineAt(prevSibling.StartPosition().Line)[:prevSibling.StartPosition().Column]

			s.AddStep("Ensure consistent indentation by using the correct spacing for each level of indentation.").
				AddFix(lib.FixSuggestion{
					NewText: "",
					StartPosition: lib.Position{
						Line: cd.MainError.Nearest.StartPosition().Line,
					},
					EndPosition: cd.MainError.Nearest.StartPosition(),
				}).
				AddFix(lib.FixSuggestion{
					NewText: spaces,
					StartPosition: lib.Position{
						Line: cd.MainError.Nearest.StartPosition().Line,
					},
					EndPosition: lib.Position{
						Line: cd.MainError.Nearest.StartPosition().Line,
					},
				})
		})
	},
}
