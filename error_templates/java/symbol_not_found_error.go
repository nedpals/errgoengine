package java

import (
	"fmt"
	"strings"

	lib "github.com/nedpals/errgoengine"
)

type symbolNotFoundErrorCtx struct {
	symbolType    string
	symbolName    string
	locationClass string
	locationNode  lib.SyntaxNode
	rootNode      lib.SyntaxNode
	parentNode    lib.SyntaxNode
}

var SymbolNotFoundError = lib.ErrorTemplate{
	Name:              "SymbolNotFoundError",
	Pattern:           comptimeErrorPattern("cannot find symbol", `symbol:\s+(?P<symbolType>variable|method|class) (?P<symbolName>\S+)\s+location\:\s+class (?P<locationClass>\S+)`),
	StackTracePattern: comptimeStackTracePattern,
	OnAnalyzeErrorFn: func(cd *lib.ContextData, m *lib.MainError) {
		symbolName := cd.Variables["symbolName"]
		errorCtx := symbolNotFoundErrorCtx{
			symbolType:    cd.Variables["symbolType"],
			locationClass: cd.Variables["locationClass"],
			symbolName:    symbolName,
			rootNode:      m.Nearest,
		}

		nodeTypeToFind := "identifier"
		if errorCtx.symbolType == "class" {
			nodeTypeToFind = "type_identifier"
		}

		for q := m.Nearest.Query("((%s) @symbol (#eq? @symbol \"%s\"))", nodeTypeToFind, symbolName); q.Next(); {
			node := q.CurrentNode()
			errorCtx.rootNode = m.Nearest
			errorCtx.parentNode = node.Parent()
			m.Nearest = node
			break
		}

		// locate the location node
		rootNode := m.Nearest.Doc.RootNode()
		for q := rootNode.Query(`(class_declaration name: (identifier) @class-name (#eq? @class-name "%s"))`, errorCtx.locationClass); q.Next(); {
			node := q.CurrentNode().Parent()
			errorCtx.locationNode = node
			break
		}

		m.Context = errorCtx
	},
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		ctx := cd.MainError.Context.(symbolNotFoundErrorCtx)
		switch ctx.symbolType {
		case "variable":
			gen.Add(`The program cannot find variable "%s"`, ctx.symbolName)
		case "method":
			gen.Add("The error indicates that the compiler cannot find the method `%s` in the `%s` class.", ctx.symbolName, ctx.locationClass)
		case "class":
			gen.Add("The error indicates that the compiler cannot find the class `%s` when attempting to create an instance of it in the `%s` class.", ctx.symbolName, ctx.locationClass)
		}
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		ctx := cd.MainError.Context.(symbolNotFoundErrorCtx)
		switch ctx.symbolType {
		case "variable":
			affectedStatementPosition := ctx.rootNode.StartPosition()
			space := getSpaceFromBeginning(cd.MainError.Document, affectedStatementPosition.Line, affectedStatementPosition.Column)

			gen.Add("Create a variable.", func(s *lib.BugFixSuggestion) {
				s.AddStep("Create a variable named \"%s\". For example:", ctx.symbolName).
					// TODO: use variable type from the inferred parameter
					AddFix(lib.FixSuggestion{
						NewText:       space + fmt.Sprintf("String %s = \"\";\n", ctx.symbolName),
						StartPosition: lib.Position{Line: affectedStatementPosition.Line, Column: 0},
						EndPosition:   lib.Position{Line: affectedStatementPosition.Line, Column: 0},
					})
			})
		case "method":
			gen.Add("Define the missing method.", func(s *lib.BugFixSuggestion) {
				bodyNode := ctx.locationNode.ChildByFieldName("body")
				lastMethodNode := bodyNode.LastNamedChild()

				methodName, parameterTypes := parseMethodSignature(ctx.symbolName)
				parameters := make([]string, len(parameterTypes))
				for i, paramType := range parameterTypes {
					parameters[i] = fmt.Sprintf("%s %c", paramType, 'a'+i) // start at a (ASCII 97)
				}

				// TODO: smartly infer the method signature for the missing method
				s.AddStep("Add the missing method `%s` to the `%s` class", methodName, ctx.locationClass).
					AddFix(lib.FixSuggestion{
						NewText:       fmt.Sprintf("\n\n\tprivate static void %s(%s) {\n\t\t// Add code here\n\t}\n", methodName, strings.Join(parameters, ", ")),
						StartPosition: lastMethodNode.EndPosition().Add(lib.Position{Column: 1}), // add 1 column so that the parenthesis won't be replaced
						EndPosition:   lastMethodNode.EndPosition().Add(lib.Position{Column: 1}), // same thing here
					})
			})
		case "class":
			gen.Add("Create the missing class", func(s *lib.BugFixSuggestion) {
				s.AddStep("Create a new class named `%s` to resolve the \"cannot find symbol\" error.", ctx.symbolName).
					AddFix(lib.FixSuggestion{
						NewText: fmt.Sprintf("class %s {\n\t// Add any necessary code for %s class\n}\n\n", ctx.symbolName, ctx.symbolName),
						StartPosition: lib.Position{
							Line:   ctx.locationNode.StartPosition().Line,
							Column: 0,
						},
						EndPosition: lib.Position{
							Line:   ctx.locationNode.StartPosition().Line,
							Column: 0,
						},
					})
			})
		}
	},
}

func parseMethodSignature(symbolName string) (methodName string, parameterTypes []string) {
	methodName = symbolName[:strings.Index(symbolName, "(")]
	parameterTypes = strings.Split(symbolName[strings.Index(symbolName, "(")+1:strings.Index(symbolName, ")")], ",")
	return methodName, parameterTypes
}
