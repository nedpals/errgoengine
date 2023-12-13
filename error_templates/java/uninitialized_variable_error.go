package java

import (
	"fmt"
	"strings"

	lib "github.com/nedpals/errgoengine"
)

type uninitializedVariableErrCtx struct {
	DeclarationSym  lib.Symbol
	DeclarationNode lib.SyntaxNode
}

var UninitializedVariableError = lib.ErrorTemplate{
	Name:              "UninitializedVariableError",
	Pattern:           comptimeErrorPattern(`variable (?P<variable>\S+) might not have been initialized`),
	StackTracePattern: comptimeStackTracePattern,
	OnAnalyzeErrorFn: func(cd *lib.ContextData, m *lib.MainError) {
		uCtx := uninitializedVariableErrCtx{}
		query := strings.NewReader(fmt.Sprintf(`((identifier) @variable (#eq? @variable "%s"))`, cd.Variables["variable"]))
		lib.QueryNode(cd.MainError.Nearest, query, func(ctx lib.QueryNodeCtx) bool {
			match := ctx.Cursor.FilterPredicates(ctx.Match, []byte(cd.MainError.Nearest.Doc.Contents))
			for _, c := range match.Captures {
				node := lib.WrapNode(cd.MainError.Nearest.Doc, c.Node)
				m.Nearest = node
				return false
			}
			return true
		})

		// get symbol and declaration node
		rootTree := lib.WrapNode(cd.MainError.Document, cd.MainError.Document.Tree.RootNode())
		nearestTree := cd.InitOrGetSymbolTree(cd.MainDocumentPath()).GetNearestScopedTree(cd.MainError.Nearest.StartPosition().Index)
		declaredVariableSym := nearestTree.GetSymbolByNode(cd.MainError.Nearest)
		declNode := rootTree.NamedDescendantForPointRange(declaredVariableSym.Location())

		uCtx.DeclarationSym = declaredVariableSym

		if !declNode.IsNull() && !declNode.Parent().IsNull() && declNode.Parent().Type() == "variable_declarator" {
			uCtx.DeclarationNode = declNode.Parent().Parent()
		}

		m.Context = uCtx
	},
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		gen.Add("This error occurs when you try to use a variable that has not been initialized with a value.")
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		ctx := cd.MainError.Context.(uninitializedVariableErrCtx)
		gen.Add("Initialize the variable", func(s *lib.BugFixSuggestion) {
			nameNode := ctx.DeclarationNode.ChildByFieldName("declarator").ChildByFieldName("name")

			s.AddStep("To resolve the uninitialized variable error, you need to initialize the `%s` variable with a value.", cd.Variables["variable"]).
				AddFix(lib.FixSuggestion{
					NewText:       fmt.Sprintf(" = %s", getDefaultValueForType(ctx.DeclarationSym.(*lib.VariableSymbol).ReturnType())),
					StartPosition: nameNode.EndPosition(),
					EndPosition:   nameNode.EndPosition(),
					Description:   "This ensures that the variable has a valid initial value before it's used.",
				})
		})

		gen.Add("Assign a value before using", func(s *lib.BugFixSuggestion) {
			fmt.Println(cd.MainError.Document.LineAt(ctx.DeclarationNode.StartPosition().Line))
			spaces := cd.MainError.Document.LineAt(ctx.DeclarationNode.StartPosition().Line)[:ctx.DeclarationNode.StartPosition().Column]

			s.AddStep("Alternatively, you can assign a value to the variable before using it.").AddFix(lib.FixSuggestion{
				NewText:       "\n" + spaces + fmt.Sprintf("%s = %s; // or any other valid value", cd.Variables["variable"], getDefaultValueForType(ctx.DeclarationSym.(*lib.VariableSymbol).ReturnType())),
				StartPosition: ctx.DeclarationNode.EndPosition(),
				EndPosition:   ctx.DeclarationNode.EndPosition(),
				Description:   "This way, the variable is initialized with a value before it's used in the statement.",
			})
		})
	},
}
