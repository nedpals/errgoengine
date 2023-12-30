package java

import (
	"fmt"

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
		q := m.Nearest.Query(`((identifier) @variable (#eq? @variable "%s"))`, cd.Variables["variable"])
		for q.Next() {
			m.Nearest = q.CurrentNode()
			break
		}

		// get symbol and declaration node
		root := m.Document.RootNode()
		nearestTree := cd.InitOrGetSymbolTree(cd.MainDocumentPath()).GetNearestScopedTree(m.Nearest.StartPosition().Index)
		declaredVariableSym := nearestTree.GetSymbolByNode(m.Nearest)
		declNode := root.NamedDescendantForPointRange(declaredVariableSym.Location())

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
