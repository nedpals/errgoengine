package java

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	lib "github.com/nedpals/errgoengine"
)

func generateVarName(inputString string) string {
	reg := regexp.MustCompile("[^a-zA-Z0-9_]")
	processedString := reg.ReplaceAllString(inputString, "")
	return strings.TrimSpace(processedString)
}

var NotAStatementError = lib.ErrorTemplate{
	Name:              "NotAStatementError",
	Pattern:           comptimeErrorPattern(`not a statement`),
	StackTracePattern: comptimeStackTracePattern,
	OnAnalyzeErrorFn: func(cd *lib.ContextData, m *lib.MainError) {
		if m.Nearest.Type() == "expression_statement" {
			m.Nearest = m.Nearest.NamedChild(0)
		}
	},
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		gen.Add("This error occurs when a line of code is written that is not a valid statement.")
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		nodeType := cd.Analyzer.AnalyzeNode(context.Background(), cd.MainError.Nearest)

		gen.Add(fmt.Sprintf("Convert the `%s` to a statement", nodeType.Name()), func(s *lib.BugFixSuggestion) {
			s.AddStep(
				"If you intended to use the `%s` as a statement, you can print it or use it in a valid statement.",
				nodeType.Name(),
			).AddFix(lib.FixSuggestion{
				NewText:       fmt.Sprintf("System.out.println(%s)", cd.MainError.Nearest.Text()),
				StartPosition: cd.MainError.Nearest.StartPosition(),
				EndPosition:   cd.MainError.Nearest.EndPosition(),
				Description:   "This change makes the string part of a valid statement by printing it to the console.",
			})
		})

		gen.Add(fmt.Sprintf("Assign the `%s` to a variable", nodeType.Name()), func(s *lib.BugFixSuggestion) {
			s.AddStep("Alternatively, you can assign the `%s` to a variable to make it a valid statement.", nodeType.Name()).
				AddFix(lib.FixSuggestion{
					NewText:       fmt.Sprintf("%s %s = %s", nodeType.Name(), generateVarName(cd.MainError.Nearest.Text()), cd.MainError.Nearest.Text()),
					StartPosition: cd.MainError.Nearest.StartPosition(),
					EndPosition:   cd.MainError.Nearest.EndPosition(),
					Description:   "This way, the string is now part of a statement and can be used later in your code.",
				})
		})
	},
}
