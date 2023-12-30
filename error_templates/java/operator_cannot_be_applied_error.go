package java

import (
	"fmt"

	lib "github.com/nedpals/errgoengine"
)

type opCannotBeAppliedCtx struct {
	Parent lib.SyntaxNode
}

var OperatorCannotBeAppliedError = lib.ErrorTemplate{
	Name:              "OperatorCannotBeAppliedError",
	Pattern:           comptimeErrorPattern("bad operand types for binary operator '(?P<operator>.)'", `first type\:\s+(?P<firstType>\S+)\s+second type\:\s+(?P<secondType>\S+)`),
	StackTracePattern: comptimeStackTracePattern,
	OnAnalyzeErrorFn: func(cd *lib.ContextData, m *lib.MainError) {
		oCtx := opCannotBeAppliedCtx{}
		operator := cd.Variables["operator"]
		for q := m.Nearest.Query(`((binary_expression) @binary_expr (#eq @binary_expr "%s"))`, operator); q.Next(); {
			node := q.CurrentNode()
			oCtx.Parent = node
			m.Nearest = node.Child(1)
			break
		}

		m.Context = oCtx
	},
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		gen.Add(
			"This error occurs when you try to apply a binary operator to incompatible operand types, such as trying to use the '%s' operator between a %s and an %s.",
			cd.Variables["operator"],
			cd.Variables["firstType"],
			cd.Variables["secondType"],
		)
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		ctx := cd.MainError.Context.(opCannotBeAppliedCtx)
		left := ctx.Parent.ChildByFieldName("left")
		right := ctx.Parent.ChildByFieldName("right")

		gen.Add(fmt.Sprintf("Use %s's compareTo method", cd.Variables["firstType"]), func(s *lib.BugFixSuggestion) {
			s.AddStep(
				"Since you are comparing a %s and an %s, you need to use the `compareTo` method to compare their values.",
				cd.Variables["firstType"],
				cd.Variables["secondType"],
			).AddFix(lib.FixSuggestion{
				NewText:       ".compareTo(String.valueOf(" + right.Text() + "))",
				StartPosition: left.EndPosition(),
				EndPosition:   left.EndPosition(),
				Description:   "The `compareTo` method returns a negative integer if the calling string is lexicographically less than the argument string.",
			})
		})

		gen.Add(fmt.Sprintf("Convert %s to %s for direct comparison", cd.Variables["secondType"], cd.Variables["firstType"]), func(s *lib.BugFixSuggestion) {
			s.AddStep(
				"If you want to compare them directly, convert the %s to %s using `%s.valueOf()`.",
				cd.Variables["secondType"],
				cd.Variables["firstType"],
				cd.Variables["firstType"],
			).AddFix(lib.FixSuggestion{
				Description:   "This ensures both operands are of the same type for comparison.",
				NewText:       ".equals(" + cd.Variables["firstType"] + ".valueOf(" + right.Text() + "))",
				StartPosition: left.EndPosition(),
				EndPosition:   ctx.Parent.EndPosition(),
			})
		})
	},
}
