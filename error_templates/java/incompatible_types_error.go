package java

import (
	"fmt"

	lib "github.com/nedpals/errgoengine"
)

type incompatibleTypesErrorCtx struct {
	Parent lib.SyntaxNode
}

var IncompatibleTypesError = lib.ErrorTemplate{
	Name:              "IncompatibleTypesError",
	Pattern:           comptimeErrorPattern(`incompatible types: (?P<leftType>\S+) cannot be converted to (?P<rightType>\S+)`),
	StackTracePattern: comptimeStackTracePattern,
	OnAnalyzeErrorFn: func(cd *lib.ContextData, m *lib.MainError) {
		iCtx := incompatibleTypesErrorCtx{}

		if m.Nearest.Type() == "expression_statement" {
			m.Nearest = m.Nearest.NamedChild(0)
		}

		if m.Nearest.Type() == "assignment_expression" {
			iCtx.Parent = m.Nearest
			m.Nearest = m.Nearest.ChildByFieldName("right")
		}

		m.Context = iCtx
	},
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		gen.Add("This error occurs when you attempt to assign a value of one data type to a variable of a different, incompatible data type.")
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		leftType := cd.Variables["leftType"]
		rightType := cd.Variables["rightType"]
		ctx := cd.MainError.Context.(incompatibleTypesErrorCtx)

		gen.Add(fmt.Sprintf("Convert %s to %s", leftType, rightType), func(s *lib.BugFixSuggestion) {
			s.AddStep("To resolve the incompatible types error, you need to explicitly convert the `%s` to a `%s`.", leftType, rightType).
				AddFix(lib.FixSuggestion{
					NewText:       fmt.Sprintf("%s.valueOf(%s)", rightType, cd.MainError.Nearest.Text()),
					StartPosition: cd.MainError.Nearest.StartPosition(),
					EndPosition:   cd.MainError.Nearest.EndPosition(),
					Description:   fmt.Sprintf("The `%s.valueOf()` method converts the `%s` to its string representation.", rightType, leftType),
				})
		})

		gen.Add(fmt.Sprintf("Concatenate %s with %s", leftType, rightType), func(s *lib.BugFixSuggestion) {
			leftVariable := ctx.Parent.ChildByFieldName("left").Text()

			s.AddStep("Alternatively, you can concatenate the `%s` with the existing `%s`.", leftType, rightType).
				AddFix(lib.FixSuggestion{
					NewText:       fmt.Sprintf("%s + ", leftVariable),
					StartPosition: cd.MainError.Nearest.StartPosition(),
					EndPosition:   cd.MainError.Nearest.StartPosition(),
					Description:   fmt.Sprintf("This converts the `%s` to a `%s` and concatenates it with the existing `%s`.", leftType, rightType, rightType),
				})
		})
	},
}
