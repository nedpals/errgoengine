package java

import (
	lib "github.com/nedpals/errgoengine"
	"github.com/nedpals/errgoengine/languages/java"
)

// const ()

// type nullPointerExceptionCtx struct {
// 	isSystemOut bool
// }

// TODO: unit testing
var NullPointerException = lib.ErrorTemplate{
	Name:    "NullPointerException",
	Pattern: runtimeErrorPattern("java.lang.NullPointerException", ""),
	// OnAnalyzeErrorFn: func(cd *lib.ContextData, m *lib.MainError) {
	// 	ctx := nullPointerExceptionCtx{}

	// 	// if the offending line is an offending method call, get the argument that triggered the null error
	// 	if cd.MainError.Nearest.Type() == "expression_statement" {
	// 		exprNode := cd.MainError.Nearest.NamedChild(0)

	// 		// NOTE: i hope we will be able to parse the entire
	// 		// java system library without hardcoding the definitions lol
	// 		if exprNode.Type() == "method_invocation" {
	// 			objNode := exprNode.ChildByFieldName("object")
	// 			// check if this is just simple printing
	// 			if objNode.Text() == "System.out" {
	// 				isSystemOut = true
	// 			} else if retType := cd.Analyzer.AnalyzeNode(exprNode); retType == java.BuiltinTypes.NullSymbol {
	// 				cd.MainError.Nearest = exprNode
	// 			}

	// 			if objNode.Type() == "array_access" {
	// 				// inArray = true
	// 				cd.MainError.Nearest = exprNode
	// 			} else {
	// 				arguments := exprNode.ChildByFieldName("arguments")
	// 				for i := 0; i < int(arguments.NamedChildCount()); i++ {
	// 					argNode := arguments.NamedChild(i)
	// 					retType := cd.Analyzer.AnalyzeNode(argNode)

	// 					if retType == java.BuiltinTypes.NullSymbol || argNode.Type() == "array_access" {
	// 						cd.MainError.Nearest = argNode
	// 						break
	// 					}
	// 				}
	// 			}
	// 		}
	// 	}
	// },
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		// TODO: create a function that will find the node with a null return type
		// sb := &strings.Builder{}
		isSystemOut := false
		// inArray := false
		methodName := ""
		origin := ""

		// if the offending line is an offending method call, get the argument that triggered the null error
		if cd.MainError.Nearest.Type() == "expression_statement" {
			exprNode := cd.MainError.Nearest.NamedChild(0)

			// NOTE: i hope we will be able to parse the entire
			// java system library without hardcoding the definitions lol
			if exprNode.Type() == "method_invocation" {
				objNode := exprNode.ChildByFieldName("object")
				// check if this is just simple printing
				if objNode.Text() == "System.out" {
					isSystemOut = true
				} else if retType := cd.Analyzer.AnalyzeNode(exprNode); retType == java.BuiltinTypes.NullSymbol {
					cd.MainError.Nearest = exprNode
				}

				if objNode.Type() == "array_access" {
					// inArray = true
					cd.MainError.Nearest = exprNode
				} else {
					arguments := exprNode.ChildByFieldName("arguments")
					for i := 0; i < int(arguments.NamedChildCount()); i++ {
						argNode := arguments.NamedChild(i)
						retType := cd.Analyzer.AnalyzeNode(argNode)

						if retType == java.BuiltinTypes.NullSymbol || argNode.Type() == "array_access" {
							cd.MainError.Nearest = argNode
							break
						}
					}
				}
			}
		}

		// identify the *MAIN* culprit
		mainNode := cd.MainError.Nearest
		switch mainNode.Type() {
		case "method_invocation":
			nameNode := mainNode.ChildByFieldName("name")
			methodName = nameNode.Text()
			origin = mainNode.ChildByFieldName("object").Text()
		default:
			origin = mainNode.Text()
		}

		if isSystemOut {
			if len(methodName) != 0 {
				gen.Add("Your program tried to print the value of \"%s\" method from \"%s\" which is a null.", methodName, origin)
			} else {
				gen.Add("Your program tried to print the value of \"%s\" which is a null", origin)
			}
		} else if len(methodName) != 0 {
			// if inArray {
			// 	gen.Add("Your program tried to execute the \"%s\" method from \"%s\" which is a null.", )
			// } else {
			gen.Add("Your program tried to execute the \"%s\" method from \"%s\" which is a null.", methodName, origin)
			// }
		}

		gen.Add("Your program try to access or manipulate an object reference that is currently pointing to `null`, meaning it doesn't refer to any actual object in memory. This typically happens when you forget to initialize an object before using it, or when you try to access an object that hasn't been properly assigned a value. ")
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		// TODO:

	},
}
