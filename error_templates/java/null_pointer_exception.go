package java

import (
	"context"

	lib "github.com/nedpals/errgoengine"
	"github.com/nedpals/errgoengine/languages/java"
)

type exceptionLocationKind int

const (
	fromUnknown          exceptionLocationKind = 0
	fromFunctionArgument exceptionLocationKind = iota
	fromSystemOut        exceptionLocationKind = iota
	fromArrayAccess      exceptionLocationKind = iota
	fromExpression       exceptionLocationKind = iota
	fromMethodInvocation exceptionLocationKind = iota
)

type nullPointerExceptionCtx struct {
	kind exceptionLocationKind
	// symbolInvolved lib.SyntaxNode
	methodName string
	origin     string
}

// TODO: unit testing
var NullPointerException = lib.ErrorTemplate{
	Name:    "NullPointerException",
	Pattern: runtimeErrorPattern("java.lang.NullPointerException", ""),
	OnAnalyzeErrorFn: func(cd *lib.ContextData, err *lib.MainError) {
		ctx := nullPointerExceptionCtx{}

		// if the offending line is an offending method call, get the argument that triggered the null error
		if cd.MainError.Nearest.Type() == "expression_statement" {
			exprNode := cd.MainError.Nearest.NamedChild(0)

			// NOTE: i hope we will be able to parse the entire
			// java system library without hardcoding the definitions lol
			if exprNode.Type() == "method_invocation" {
				objNode := exprNode.ChildByFieldName("object")
				// check if this is just simple printing
				if objNode.Text() == "System.out" {
					ctx.kind = fromSystemOut
				} else if retType := cd.Analyzer.AnalyzeNode(context.Background(), exprNode); retType == java.BuiltinTypes.NullSymbol {
					cd.MainError.Nearest = exprNode
					ctx.kind = fromMethodInvocation
				}

				if objNode.Type() == "array_access" {
					// inArray = true
					cd.MainError.Nearest = exprNode
					ctx.kind = fromArrayAccess
				} else {
					arguments := exprNode.ChildByFieldName("arguments")
					for i := 0; i < int(arguments.NamedChildCount()); i++ {
						argNode := arguments.NamedChild(i)
						retType := cd.Analyzer.AnalyzeNode(context.Background(), argNode)

						if retType == java.BuiltinTypes.NullSymbol || argNode.Type() == "array_access" {
							cd.MainError.Nearest = argNode
							ctx.kind = fromFunctionArgument
							break
						}
					}
				}
			} else if exprNode.Type() == "assignment_expression" {
				// right := exprNode.ChildByFieldName("right")
				//
			}

			// identify the *MAIN* culprit
			mainNode := cd.MainError.Nearest
			switch mainNode.Type() {
			case "method_invocation":
				nameNode := mainNode.ChildByFieldName("name")
				ctx.methodName = nameNode.Text()
				ctx.origin = mainNode.ChildByFieldName("object").Text()
			default:
				ctx.origin = mainNode.Text()
			}

			err.Context = ctx
		}
	},
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		// TODO: create a function that will find the node with a null return type
		ctx := cd.MainError.Context.(nullPointerExceptionCtx)

		if ctx.kind == fromSystemOut {
			gen.Add("Your program tried to print the value of ")
			if len(ctx.methodName) != 0 {
				gen.Add("\"%s\" method from ", ctx.methodName)
			}
			gen.Add("\"%s\" which is a null.", ctx.origin)
			return
		} else if len(ctx.methodName) != 0 {
			// if inArray {
			// 	gen.Add("Your program tried to execute the \"%s\" method from \"%s\" which is a null.", )
			// } else {
			gen.Add("Your program tried to execute the \"%s\" method from \"%s\" which is a null.", ctx.methodName, ctx.origin)
			// }
			return
		}

		gen.Add("Your program try to access or manipulate an object reference that is currently pointing to `null`, meaning it doesn't refer to any actual object in memory. This typically happens when you forget to initialize an object before using it, or when you try to access an object that hasn't been properly assigned a value. ")
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		gen.Add("Wrap with an if statement", func(s *lib.BugFixSuggestion) {
			s.AddDescription("Check for the variable that is being used as `null`.")
		})

		gen.Add("Initialize the variable", func(s *lib.BugFixSuggestion) {
			s.AddDescription("An alternative fix is to initialize the `test` variable with a non-null value before calling the method.")
		})
	},
}
