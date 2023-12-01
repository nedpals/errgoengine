package java

import (
	"strings"

	lib "github.com/nedpals/errgoengine"
)

type arithExceptionKind int

const (
	unknown               arithExceptionKind = 0
	dividedByZero         arithExceptionKind = iota
	nonTerminatingDecimal arithExceptionKind = iota
)

type arithExceptionCtx struct {
	kind arithExceptionKind
}

var ArithmeticException = lib.ErrorTemplate{
	Name:    "ArithmeticException",
	Pattern: runtimeErrorPattern("java.lang.ArithmeticException", "(?P<reason>.+)"),
	OnAnalyzeErrorFn: func(cd *lib.ContextData, err *lib.MainError) {
		ctx := arithExceptionCtx{}
		reason := cd.Variables["reason"]
		query := ""

		switch reason {
		case "/ by zero":
			ctx.kind = dividedByZero
			query = "(_) \"/\" ((decimal_integer_literal) @literal (#eq? @literal \"0\"))"
		case "Non-terminating decimal expansion; no exact representable decimal result.":
			ctx.kind = nonTerminatingDecimal
			query = "(method_invocation) @methodCall (#eq? @methodCall \".divide\")"
		default:
			ctx.kind = unknown
		}

		if len(query) != 0 {
			lib.QueryNode(cd.MainError.Nearest, strings.NewReader(query), func(ctx lib.QueryNodeCtx) bool {
				match := ctx.Cursor.FilterPredicates(ctx.Match, []byte(cd.MainError.Nearest.Doc.Contents))
				for _, c := range match.Captures {
					node := lib.WrapNode(cd.MainError.Nearest.Doc, c.Node)
					err.Nearest = node
					return false
				}
				return true
			})
		}

		err.Context = ctx
	},
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		ctx := cd.MainError.Context.(arithExceptionCtx)
		switch ctx.kind {
		case dividedByZero:
			gen.Add("This error is raised when you try to perform arithmetic operations that are not mathematically possible, such as division by zero.")
		case nonTerminatingDecimal:
			gen.Add("This error is raised when dividing two `BigDecimal` numbers, and the division operation results in a non-terminating decimal expansion, meaning the division produces a non-repeating and non-terminating decimal.")
		case unknown:
			gen.Add("You just encountered an unknown `ArithmeticException` error of which we cannot explain to you properly.")
		}
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		ctx := cd.MainError.Context.(arithExceptionCtx)
		switch ctx.kind {
		case dividedByZero:
			gen.Add("Avoid dividing by zero.", func(s *lib.BugFixSuggestion) {
				s.AddStep("To fix the 'ArithmeticException: / by zero', you need to ensure you are not dividing by zero, which is mathematically undefined.").
					AddFix(lib.FixSuggestion{
						NewText:       "1",
						Description:   "This adjustment replaces the division by zero with a value that is not zero, ensuring the operation is valid. Division by zero is mathematically undefined, causing an 'ArithmeticException'. By changing the denominator to a non-zero value, you prevent the error.",
						StartPosition: cd.MainError.Nearest.StartPosition(),
						EndPosition:   cd.MainError.Nearest.EndPosition(),
					})
			})
		case nonTerminatingDecimal:
			gen.Add("Ensure precise division", func(s *lib.BugFixSuggestion) {
				s.AddStep("To fix the 'ArithmeticException: Non-terminating decimal expansion', you need to ensure the division operation is precise.").
					AddFix(lib.FixSuggestion{
						NewText:       ", RoundingMode.HALF_UP)",
						StartPosition: cd.MainError.Nearest.EndPosition(),
						EndPosition:   cd.MainError.Nearest.EndPosition(),
					})
			})

			if parent := cd.MainError.Nearest.Parent(); parent.Type() == "block" {
				gen.Add("Catch ArithmeticException", func(s *lib.BugFixSuggestion) {
					firstChild := parent.FirstNamedChild()
					lastChild := parent.LastNamedChild()

					s.AddStep("Handle the ArithmeticException by wrapping the division operation in a try-catch block to manage the potential exception and inform the user about the non-terminating result.").
						AddFix(lib.FixSuggestion{
							NewText:       "try {",
							StartPosition: firstChild.StartPosition(),
							EndPosition:   firstChild.StartPosition(),
						}).
						AddFix(lib.FixSuggestion{
							NewText:       "} catch (ArithmeticException e) {\n\tSystem.out.println(\"Non-terminating result: \" + e.getMessage());\n}",
							StartPosition: lastChild.StartPosition(),
							EndPosition:   lastChild.StartPosition(),
						})
				})
			}
		}
	},
}
