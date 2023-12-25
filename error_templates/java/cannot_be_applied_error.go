package java

import (
	"fmt"
	"strings"

	lib "github.com/nedpals/errgoengine"
	"github.com/nedpals/errgoengine/utils/numbers"
)

type cannotBeAppliedErrorKind int

const (
	cannotBeAppliedMismatchedArgType  cannotBeAppliedErrorKind = 0
	cannotBeAppliedMismatchedArgCount cannotBeAppliedErrorKind = iota
)

type cannotBeAppliedErrorCtx struct {
	rawRequiredTypes []string
	rawFoundTypes    []string
	requiredTypes    []lib.Symbol
	foundTypes       []lib.Symbol
	callExprNode     lib.SyntaxNode
	kind             cannotBeAppliedErrorKind
	invalidIdx       int
}

var CannotBeAppliedError = lib.ErrorTemplate{
	Name: "CannotBeAppliedError",
	Pattern: comptimeErrorPattern(
		`method (?P<method>\S+) in class (?P<className>\S+) cannot be applied to given types;`,
		`required:\s+(?P<requiredTypes>.+)\s+found:\s+(?P<foundTypes>.+)\s+reason:\s+(?P<reason>.+)`,
	),
	StackTracePattern: comptimeStackTracePattern,
	OnAnalyzeErrorFn: func(cd *lib.ContextData, m *lib.MainError) {
		cCtx := cannotBeAppliedErrorCtx{
			rawRequiredTypes: strings.Split(cd.Variables["requiredTypes"], ","),
			rawFoundTypes:    strings.Split(cd.Variables["foundTypes"], ","),
			requiredTypes:    []lib.Symbol{},
			foundTypes:       []lib.Symbol{},
		}

		// get the required types
		for _, rawType := range cCtx.rawRequiredTypes {
			sym := cd.FindSymbol(rawType, 0)
			cCtx.requiredTypes = append(cCtx.requiredTypes, sym)
		}

		// get the found types
		for _, rawType := range cCtx.rawFoundTypes {
			sym := cd.FindSymbol(rawType, 0)
			cCtx.foundTypes = append(cCtx.foundTypes, sym)
		}

		// get invalid idx
		if len(cCtx.rawFoundTypes) > len(cCtx.rawRequiredTypes) {
			cCtx.kind = cannotBeAppliedMismatchedArgCount
			cCtx.invalidIdx = len(cCtx.rawFoundTypes) - 1
		} else {
			cCtx.kind = cannotBeAppliedMismatchedArgType

			for i := 0; i < len(cCtx.requiredTypes); i++ {
				if i >= len(cCtx.foundTypes) {
					break
				}

				if cCtx.requiredTypes[i] != cCtx.foundTypes[i] {
					cCtx.invalidIdx = i
					break
				}
			}
		}

		// query nearest node
		argumentNodeTypesToLook := ""
		for _, sym := range cCtx.foundTypes {
			valueNodeTypes := symbolToValueNodeType(sym)
			nTypesStr := "[(identifier) (field_access)"

			for _, nType := range valueNodeTypes {
				nTypesStr += fmt.Sprintf(" (%s)", nType)
			}

			nTypesStr += "]"
			argumentNodeTypesToLook += nTypesStr
		}

		rawQuery := fmt.Sprintf(`((method_invocation name: (identifier) @name arguments: (argument_list %s)) @call (#eq? @name "%s"))`, argumentNodeTypesToLook, cd.Variables["method"])

		lib.QueryNode(m.Nearest, strings.NewReader(rawQuery), func(ctx lib.QueryNodeCtx) bool {
			match := ctx.Cursor.FilterPredicates(ctx.Match, []byte(m.Nearest.Doc.Contents))
			for _, c := range match.Captures {
				node := lib.WrapNode(m.Document, c.Node)
				fmt.Println(node.Text())
				cCtx.callExprNode = node
				argNode := node.ChildByFieldName("arguments").NamedChild(cCtx.invalidIdx)
				m.Nearest = argNode
				return false
			}
			return true
		})

		m.Context = cCtx
	},
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		ctx := cd.MainError.Context.(cannotBeAppliedErrorCtx)

		switch ctx.kind {
		case cannotBeAppliedMismatchedArgCount:
			gen.Add("This error occurs when there is an attempt to apply a method with an incorrect number of arguments.")
		case cannotBeAppliedMismatchedArgType:
			gen.Add("This error occurs when there is an attempt to apply a method with arguments that do not match the method signature.")
		default:
			gen.Add("unable to determine.")
		}
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		ctx := cd.MainError.Context.(cannotBeAppliedErrorCtx)

		switch ctx.kind {
		case cannotBeAppliedMismatchedArgCount:
			gen.Add("Use the correct number of arguments", func(s *lib.BugFixSuggestion) {
				s.AddStep("Modify the `%s` method call to use only %s argument.", cd.Variables["method"], numbers.ToWords(len(ctx.rawRequiredTypes))).
					AddFix(lib.FixSuggestion{
						NewText:       "",
						StartPosition: cd.MainError.Nearest.PrevSibling().StartPosition(),
						EndPosition:   cd.MainError.Nearest.EndPosition(),
					})
			})
		case cannotBeAppliedMismatchedArgType:
			gen.Add("Use the correct argument types", func(s *lib.BugFixSuggestion) {
				s.AddStep("Provide the correct argument types when calling the `%s` method", cd.Variables["method"]).
					AddFix(lib.FixSuggestion{
						NewText:       castValueNode(cd.MainError.Nearest, ctx.requiredTypes[ctx.invalidIdx]),
						StartPosition: cd.MainError.Nearest.StartPosition(),
						EndPosition:   cd.MainError.Nearest.EndPosition(),
					})
			})
		}

	},
}
