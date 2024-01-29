package python

import (
	"fmt"

	lib "github.com/nedpals/errgoengine"
)

var charToWord = map[string]string{
	"(": "open parenthesis",
}

var charPairs = map[string]string{
	"(": ")",
}

var SyntaxError = lib.ErrorTemplate{
	Name:    "SyntaxError",
	Pattern: compileTimeError("SyntaxError: '(?P<character>.+)' was never closed"),
	OnAnalyzeErrorFn: func(cd *lib.ContextData, m *lib.MainError) {
		for q := m.Nearest.Query(`(ERROR) @err`); q.Next(); {
			m.Nearest = q.CurrentNode()
			break
		}
	},
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		text := fmt.Sprintf("the '%s'", cd.Variables["character"])
		if word, ok := charToWord[cd.Variables["character"]]; ok {
			text = fmt.Sprintf("the %s `%s`", word, cd.Variables["character"])
		}

		gen.Add(
			"This error occurs when there is a syntax error in the code, and %s is not closed properly.",
			text)
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {

		if pair, ok := charPairs[cd.Variables["character"]]; ok {
			text := fmt.Sprintf("the `%s`", cd.Variables["character"])
			if word, ok := charToWord[cd.Variables["character"]]; ok {
				text = fmt.Sprintf("the %s", word)
			}

			gen.Add("Close "+text, func(s *lib.BugFixSuggestion) {
				text = fmt.Sprintf("%s (`%s`)", text, cd.Variables["character"])

				s.AddStep("Ensure that %s is closed properly.", text).
					AddFix(lib.FixSuggestion{
						NewText:       pair,
						StartPosition: cd.MainError.Nearest.EndPosition(),
						EndPosition:   cd.MainError.Nearest.EndPosition(),
					})
			})
		}
	},
}
