package java

import (
	"fmt"
	"strings"

	lib "github.com/nedpals/errgoengine"
	"github.com/nedpals/errgoengine/languages/java"
)

type exceptionLocationKind int

const (
	fromUnknown exceptionLocationKind = iota
	fromFunctionArgument
	fromSystemOut
	fromArrayAccess
	fromExpression
	fromMethodInvocation
	fromAssignment
)

type nullPointerExceptionCtx struct {
	kind exceptionLocationKind
	// symbolInvolved lib.SyntaxNode
	methodName string
	origin     string
	parent     lib.SyntaxNode
}

// TODO: unit testing
var NullPointerException = lib.ErrorTemplate{
	Name:    "NullPointerException",
	Pattern: runtimeErrorPattern("java.lang.NullPointerException", ""),
	OnAnalyzeErrorFn: func(cd *lib.ContextData, m *lib.MainError) {
		ctx := nullPointerExceptionCtx{}

		// NOTE: i hope we will be able to parse the entire
		// java system library without hardcoding the definitions lol
		for q := m.Nearest.Query(`([
			(field_access object: (_) @ident field: (identifier))
			(method_invocation object: [
				(identifier) @ident
				(field_access object: (_) @obj field: (identifier)) @ident
			]) @call
			(method_invocation arguments: (argument_list (identifier) @ident))
			(array_access) @access
		]
			(#any-match? @obj "^[a-z0-9_]+")
			(#not-match? @ident "^[A-Z][a-z0-9_]$"))`); q.Next(); {
			tagName := q.CurrentTagName()
			node := q.CurrentNode()

			if tagName == "call" {
				if strings.HasPrefix(node.Text(), "System.out.") {
					ctx.kind = fromSystemOut
				} else {
					// use the next node tagged with "ident"
					continue
				}
			} else if tagName == "access" {
				cd.MainError.Nearest = node
				ctx.kind = fromArrayAccess
			} else if tagName == "ident" {
				retType := lib.UnwrapActualReturnType(cd.FindSymbol(node.Text(), node.StartPosition().Index))
				if retType != java.BuiltinTypes.NullSymbol {
					continue
				}

				ctx.origin = node.Text()
				parent := node.Parent()
				switch parent.Type() {
				case "field_access":
					ctx.kind = fromExpression
				case "argument_list":
					// if the offending line is an offending method call, get the argument that triggered the null error
					ctx.kind = fromFunctionArgument
				case "method_invocation":
					ctx.kind = fromMethodInvocation
					ctx.methodName = parent.ChildByFieldName("name").Text()
					ctx.origin = parent.ChildByFieldName("object").Text()
				case "assignment_expression":
					ctx.kind = fromAssignment
				}

				ctx.parent = parent
				m.Nearest = node
				break
			}
		}

		m.Context = ctx
	},
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		// TODO: create a function that will find the node with a null return type
		ctx := cd.MainError.Context.(nullPointerExceptionCtx)

		if ctx.kind == fromSystemOut {
			gen.Add("The error occurs due to your program tried to print the value of ")
			if len(ctx.methodName) != 0 {
				gen.Add("\"%s\" method from ", ctx.methodName)
			}
			gen.Add("\"%s\" which is a null.", ctx.origin)
			return
		} else if len(ctx.methodName) != 0 {
			// if inArray {
			// 	gen.Add("Your program tried to execute the \"%s\" method from \"%s\" which is a null.", )
			// } else {
			gen.Add("The error occurs due to your program tried to execute the \"%s\" method from \"%s\" which is a null.", ctx.methodName, ctx.origin)
			// }
			return
		}

		gen.Add("Your program try to access or manipulate an object reference that is currently pointing to `null`, meaning it doesn't refer to any actual object in memory. This typically happens when you forget to initialize an object before using it, or when you try to access an object that hasn't been properly assigned a value. ")
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		ctx := cd.MainError.Context.(nullPointerExceptionCtx)
		parent := ctx.parent
		for parent.Type() != "expression_statement" {
			if parent.Parent().IsNull() {
				break
			}
			parent = parent.Parent()
		}

		if parent.Type() == "expression_statement" {
			spaces := cd.MainError.Document.LineAt(parent.StartPosition().Line)[:parent.StartPosition().Column]

			gen.Add("Wrap with an if statement", func(s *lib.BugFixSuggestion) {
				s.AddStep("Check for the variable that is being used as `null`.").
					AddFix(lib.FixSuggestion{
						NewText:       fmt.Sprintf("if (%s != null) {\n", ctx.origin) + strings.Repeat(spaces, 2),
						StartPosition: parent.StartPosition(),
						EndPosition:   parent.StartPosition(),
					}).
					AddFix(lib.FixSuggestion{
						NewText:       "\n" + spaces + "}\n",
						StartPosition: parent.EndPosition(),
						EndPosition:   parent.EndPosition(),
					})
			})
		}

		gen.Add("Initialize the variable", func(s *lib.BugFixSuggestion) {
			// get the original location of variable
			symbolTree := cd.InitOrGetSymbolTree(cd.MainDocumentPath())
			varSym := symbolTree.GetSymbolByNode(cd.MainError.Nearest)

			loc := varSym.Location()
			varDeclNode := cd.MainError.Document.RootNode().NamedDescendantForPointRange(loc)
			if varDeclNode.Type() == "variable_declarator" {
				loc = varDeclNode.ChildByFieldName("value").Location()
			}

			s.AddStep("An alternative fix is to initialize the `%s` variable with a non-null value before calling the method.", ctx.origin).
				AddFix(lib.FixSuggestion{
					NewText:       getDefaultValueForType(lib.UnwrapReturnType(varSym)),
					StartPosition: loc.StartPos,
					EndPosition:   loc.EndPos,
				})
		})
	},
}
