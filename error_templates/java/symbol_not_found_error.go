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

		query := fmt.Sprintf("((identifier) @symbol (#eq? @symbol \"%s\"))", symbolName)
		lib.QueryNode(m.Nearest, strings.NewReader(query), func(ctx lib.QueryNodeCtx) bool {
			match := ctx.Cursor.FilterPredicates(ctx.Match, []byte(m.Nearest.Doc.Contents))
			for _, c := range match.Captures {
				node := lib.WrapNode(m.Nearest.Doc, c.Node)
				errorCtx.rootNode = m.Nearest
				errorCtx.parentNode = node.Parent()
				m.Nearest = node
				return false
			}
			return true
		})

		// locate the location node
		locationQuery := fmt.Sprintf(`(class_declaration name: (identifier) @class-name (#eq? @class-name "%s"))`, errorCtx.locationClass)
		rootNode := lib.WrapNode(m.Nearest.Doc, m.Nearest.Doc.Tree.RootNode())

		lib.QueryNode(rootNode, strings.NewReader(locationQuery), func(ctx lib.QueryNodeCtx) bool {
			match := ctx.Cursor.FilterPredicates(ctx.Match, []byte(m.Nearest.Doc.Contents))
			for _, c := range match.Captures {
				node := lib.WrapNode(m.Nearest.Doc, c.Node.Parent())
				errorCtx.locationNode = node
				return false
			}
			return true
		})

		m.Context = errorCtx
	},
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		ctx := cd.MainError.Context.(symbolNotFoundErrorCtx)
		switch ctx.symbolType {
		case "variable":
			gen.Add(`The program cannot find variable "%s"`, ctx.symbolName)
		case "method":
			gen.Add("The error indicates that the compiler cannot find the method `%s` in the `%s` class.", ctx.symbolName, ctx.locationClass)
		}
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		ctx := cd.MainError.Context.(symbolNotFoundErrorCtx)
		switch ctx.symbolType {
		case "variable":
			gen.Add("Create a variable.", func(s *lib.BugFixSuggestion) {
				s.AddStep("Create a variable named \"%s\". For example:", ctx.symbolName).
					// TODO: use variable type from the inferred parameter
					AddFix(lib.FixSuggestion{
						NewText:       fmt.Sprintf("String %s = \"\";\n", ctx.symbolName),
						StartPosition: ctx.rootNode.StartPosition(),
						EndPosition:   ctx.rootNode.StartPosition(),
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
		}
	},
}

func parseMethodSignature(symbolName string) (methodName string, parameterTypes []string) {
	methodName = symbolName[:strings.Index(symbolName, "(")]
	parameterTypes = strings.Split(symbolName[strings.Index(symbolName, "(")+1:strings.Index(symbolName, ")")], ",")
	return methodName, parameterTypes
}
