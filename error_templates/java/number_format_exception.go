package java

import (
	"context"

	lib "github.com/nedpals/errgoengine"
)

var NumberFormatException = lib.ErrorTemplate{
	Name:    "NumberFormatException",
	Pattern: runtimeErrorPattern("java.lang.NumberFormatException", "For input string: \"(?P<string>.+)\""),
	OnAnalyzeErrorFn: func(cd *lib.ContextData, m *lib.MainError) {
		for q := m.Document.RootNode().Query(`(method_invocation name: (identifier) @name arguments: (argument_list (_) @arg) (#eq? @name "parseInt") (#any-eq? @arg "%s"))`, cd.Variables["string"]); q.Next(); {
			if q.CurrentTagName() != "arg" {
				continue
			}
			node := q.CurrentNode()
			m.Nearest = node
			break
		}
	},
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		gen.Add("This error occurs when there is an attempt to convert a string to a numeric type, but the string does not represent a valid number.")
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		gen.Add("Ensure valid input for parsing", func(s *lib.BugFixSuggestion) {
			if cd.MainError.Nearest.Type() == "identifier" {
				variableSym := cd.Analyzer.AnalyzeNode(context.TODO(), cd.MainError.Nearest)

				if variableSym != nil {
					varNode := cd.MainError.Document.RootNode().NamedDescendantForPointRange(
						variableSym.Location(),
					)

					if varNode.Type() != "variable_declarator" {
						return
					}

					valueNode := varNode.ChildByFieldName("value")
					if valueNode.Type() != "string_literal" {
						return
					}

					s.AddStep("Make sure the string contains a valid numeric representation before attempting to parse it.").
						AddFix(lib.FixSuggestion{
							NewText:       "123",
							StartPosition: varNode.ChildByFieldName("value").StartPosition().Add(lib.Position{Column: 1}),
							EndPosition:   varNode.ChildByFieldName("value").EndPosition().Add(lib.Position{Column: -1}),
						})
				}
			}
		})
	},
}
