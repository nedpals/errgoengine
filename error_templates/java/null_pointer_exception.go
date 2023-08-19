package java

import (
	"fmt"

	"github.com/nedpals/errgoengine/languages/java"
	"github.com/nedpals/errgoengine/lib"
)

// TODO: unit testing
var NullPointerException = lib.ErrorTemplate{
	Name:    "NullPointerException",
	Pattern: runtimeErrorPattern("java.lang.NullPointerException", ""),
	OnGenExplainFn: func(cd *lib.ContextData) string {
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
				} else if retType := cd.AnalyzeValue(exprNode); retType == java.BuiltinTypes.NullSymbol {
					cd.MainError.Nearest = exprNode
				}

				if objNode.Type() == "array_access" {
					// inArray = true
					cd.MainError.Nearest = exprNode
				} else {
					arguments := exprNode.ChildByFieldName("arguments")
					for i := 0; i < int(arguments.NamedChildCount()); i++ {
						argNode := arguments.NamedChild(i)
						retType := cd.AnalyzeValue(argNode)

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
				return fmt.Sprintf("Your program tried to print the value of \"%s\" method from \"%s\" which is a null.", methodName, origin)
			} else {
				return fmt.Sprintf("Your program tried to print the value of \"%s\" which is a null", origin)
			}
		} else if len(methodName) != 0 {
			// if inArray {
			// 	return fmt.Sprintf("Your program tried to execute the \"%s\" method from \"%s\" which is a null.", )
			// } else {
			return fmt.Sprintf("Your program tried to execute the \"%s\" method from \"%s\" which is a null.", methodName, origin)
			// }
		}

		return "Your program try to access or manipulate an object reference that is currently pointing to `null`, meaning it doesn't refer to any actual object in memory. This typically happens when you forget to initialize an object before using it, or when you try to access an object that hasn't been properly assigned a value. "
	},
	OnGenBugFixFn: func(cd *lib.ContextData) []lib.BugFix {
		// TODO:
		return []lib.BugFix{}
	},
}
