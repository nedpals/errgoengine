package java

import (
	"fmt"
	"strings"

	lib "github.com/nedpals/errgoengine"
)

type symbolNotFoundErrorCtx struct {
	symbolType       string
	symbolName       string
	locationClass    string
	locationVariable string
	locationVarType  string
	variableType     lib.Symbol // only if the locationVariable is present
	classLocation    lib.Location
	locationNode     lib.SyntaxNode
	rootNode         lib.SyntaxNode
	parentNode       lib.SyntaxNode
}

var SymbolNotFoundError = lib.ErrorTemplate{
	Name:              "SymbolNotFoundError",
	Pattern:           comptimeErrorPattern("cannot find symbol", `symbol:\s+(?P<symbolType>variable|method|class) (?P<symbolName>\S+)\s+location\:\s+(?:(?:class (?P<locationClass>\S+))|(?:variable (?P<locationVariable>\S+) of type (?P<locationVarType>\S+)))`),
	StackTracePattern: comptimeStackTracePattern,
	OnAnalyzeErrorFn: func(cd *lib.ContextData, m *lib.MainError) {
		symbolName := cd.Variables["symbolName"]
		errorCtx := symbolNotFoundErrorCtx{
			symbolType:       cd.Variables["symbolType"],
			locationClass:    cd.Variables["locationClass"],
			locationVariable: cd.Variables["locationVariable"],
			locationVarType:  cd.Variables["locationVarType"],
			symbolName:       symbolName,
			variableType:     lib.UnresolvedSymbol,
			rootNode:         m.Nearest,
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
		if len(errorCtx.locationClass) > 0 {
			rootNode := m.Nearest.Doc.RootNode()
			for q := rootNode.Query(`(class_declaration name: (identifier) @class-name (#eq? @class-name "%s"))`, errorCtx.locationClass); q.Next(); {
				node := q.CurrentNode().Parent()
				errorCtx.locationNode = node
				break
			}
		} else if len(errorCtx.locationVariable) > 0 && len(errorCtx.locationVarType) > 0 {
			rootNode := m.Nearest.Doc.RootNode()
			for q := rootNode.Query(`
				(local_variable_declaration
					type: (_) @var-type
					declarator: (variable_declarator
						name: (identifier) @var-name
						(#eq? @var-name "%s"))
					(#eq? @var-type "%s"))`,
				errorCtx.locationVariable,
				errorCtx.locationVarType); q.Next(); q.Next() {
				node := q.CurrentNode().Parent()
				errorCtx.locationNode = node
				break
			}
		} else if len(errorCtx.locationVariable) > 0 {
			rootNode := m.Nearest.Doc.RootNode()
			for q := rootNode.Query(`(variable_declarator name: (identifier) @var-name (#eq? @var-name "%s"))`, errorCtx.locationVariable); q.Next(); {
				node := q.CurrentNode().Parent()
				errorCtx.locationNode = node
				break
			}

			// get type of the variable
			declNode := errorCtx.locationNode.Parent() // assumes that this is local_variable_declaration
			typeNode := declNode.ChildByFieldName("type")

			errorCtx.locationVarType = typeNode.Text()
		}

		if len(errorCtx.locationVarType) != 0 {
			errorCtx.locationClass = errorCtx.locationVarType
			foundVariableType := cd.FindSymbol(errorCtx.locationVarType, -1)
			if foundVariableType != nil {
				errorCtx.variableType = foundVariableType

				// cast to top level symbol
				if topLevelSym, ok := foundVariableType.(*lib.TopLevelSymbol); ok {
					errorCtx.classLocation = topLevelSym.Location()
				}
			} else {
				errorCtx.variableType = lib.UnresolvedSymbol
			}
		}

		if !errorCtx.locationNode.IsNull() && errorCtx.locationNode.Type() != "class_declaration" {
			if errorCtx.classLocation.DocumentPath != "" {
				// go to the class declaration of that specific class
				doc := cd.Store.Documents[errorCtx.classLocation.DocumentPath]
				foundNode := doc.RootNode().NamedDescendantForPointRange(errorCtx.classLocation)

				if !foundNode.IsNull() {
					errorCtx.locationNode = foundNode
				}
			} else {
				// go up to the class declaration
				for !errorCtx.locationNode.IsNull() && errorCtx.locationNode.Type() != "class_declaration" {
					errorCtx.locationNode = errorCtx.locationNode.Parent()
				}
			}
		}

		m.Context = errorCtx
	},
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		ctx := cd.MainError.Context.(symbolNotFoundErrorCtx)
		switch ctx.symbolType {
		case "variable":
			gen.Add(`The error indicates that the compiler cannot find variable "%s"`, ctx.symbolName)
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
			// change the doc to use for defining the missing method
			if ctx.classLocation.DocumentPath != "" {
				gen.Document = cd.Store.Documents[ctx.classLocation.DocumentPath]
			}

			gen.Add("Define the missing method.", func(s *lib.BugFixSuggestion) {
				bodyNode := ctx.locationNode.ChildByFieldName("body")
				lastMethodNode := bodyNode.LastNamedChild()

				methodName, parameterTypes := parseMethodSignature(ctx.symbolName)
				parameters := make([]string, len(parameterTypes))
				for i, paramType := range parameterTypes {
					parameters[i] = fmt.Sprintf("%s %c", paramType, 'a'+i) // start at a (ASCII 97)
				}

				// TODO: smartly infer the method signature for the missing method
				prefix := "Add"
				if ctx.classLocation.DocumentPath != "" {
					prefix = fmt.Sprintf("In `%s`, add", ctx.classLocation.DocumentPath)
				}

				s.AddStep("%s the missing method `%s` to the `%s` class", prefix, methodName, ctx.locationClass).
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
	openingPar := strings.Index(symbolName, "(")
	closingPar := strings.Index(symbolName, ")")

	methodName = symbolName[:openingPar]
	if openingPar+1 == closingPar {
		return
	}

	parameterTypes = strings.Split(symbolName[openingPar+1:closingPar], ",")
	return methodName, parameterTypes
}
