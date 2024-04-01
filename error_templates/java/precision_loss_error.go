package java

import (
	"fmt"

	lib "github.com/nedpals/errgoengine"
)

type precisionLossCtx struct {
	Parent lib.SyntaxNode
}

var PrecisionLossError = lib.ErrorTemplate{
	Name:              "PrecisionLossError",
	Pattern:           comptimeErrorPattern(`incompatible types: possible lossy conversion from (?P<currentType>\S+) to (?P<targetType>\S+)`),
	StackTracePattern: comptimeStackTracePattern,
	OnAnalyzeErrorFn: func(cd *lib.ContextData, m *lib.MainError) {
		pCtx := precisionLossCtx{}
		targetType := cd.Variables["targetType"]
		for q := m.Nearest.Query(`((local_variable_declaration type: (_) @target-type) (#eq? @target-type "%s"))`, targetType); q.Next(); {
			node := q.CurrentNode()
			pCtx.Parent = node.Parent()
			m.Nearest = pCtx.Parent.ChildByFieldName("declarator").ChildByFieldName("value")
			break
		}

		m.Context = pCtx
	},
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		gen.Add(
			"This error occurs when you try to assign a value from a data type with higher precision (%s) to a data type with lower precision (%s), which may result in a loss of precision.",
			cd.Variables["currentType"],
			cd.Variables["targetType"],
		)
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		// ctx := cd.MainError.Context.(precisionLossCtx)
		variableInvolved := cd.MainError.Nearest
		gen.Add(fmt.Sprintf("Explicitly cast to %s", cd.Variables["targetType"]), func(s *lib.BugFixSuggestion) {
			s.AddStep("To resolve the precision loss, explicitly cast the `%s` to %s.", variableInvolved.Text(), cd.Variables["targetType"]).AddFix(lib.FixSuggestion{
				NewText:       fmt.Sprintf("(%s) ", cd.Variables["targetType"]),
				StartPosition: variableInvolved.StartPosition(),
				EndPosition:   variableInvolved.StartPosition(),
				Description:   "This casting informs the compiler about the potential loss of precision and allows the assignment.",
			})
		})

		if cd.Variables["targetType"] != "int" {
			gen.Add(fmt.Sprintf("Use an 'f' suffix for the %s literal", cd.Variables["targetType"]), func(s *lib.BugFixSuggestion) {
				nearestTree := cd.InitOrGetSymbolTree(cd.MainError.DocumentPath()).GetNearestScopedTree(variableInvolved.StartPosition().Index)

				involvedVariable := nearestTree.GetSymbolByNode(variableInvolved)
				if involvedVariable == nil {
					// TODO: remove this check(?)
					return
				}

				node := cd.MainError.Document.RootNode().
					NamedDescendantForPointRange(involvedVariable.Location())

				involvedVariableValueNode := node.ChildByFieldName("value")

				s.AddStep(
					"Alternatively, you can use the 'f' suffix to specify that the literal is of type %s.",
					cd.Variables["targetType"]).AddFix(lib.FixSuggestion{
					NewText:       involvedVariableValueNode.Text() + "f",
					StartPosition: variableInvolved.StartPosition(),
					EndPosition:   variableInvolved.EndPosition(),
					Description:   "This way, you directly define the float variable without the need for casting.",
				})
			})
		}
	},
}
