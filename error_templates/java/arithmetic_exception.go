package java

import lib "github.com/nedpals/errgoengine"

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
		switch reason {
		case "/ by zero":
			ctx.kind = dividedByZero
		case "Non-terminating decimal expansion; no exact representable decimal result.":
			ctx.kind = nonTerminatingDecimal
		default:
			ctx.kind = unknown
		}
		err.Context = ctx
	},
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		ctx := cd.MainError.Context.(arithExceptionCtx)
		switch ctx.kind {
		case dividedByZero:
			ctx.kind = dividedByZero
			gen.Add("One of your variables initialized a double value by dividing a number to zero")
		case nonTerminatingDecimal:
			gen.Add("TODO")
		case unknown:
			gen.Add("Unknown ArithmeticException")
		}
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		// TODO:
	},
}
