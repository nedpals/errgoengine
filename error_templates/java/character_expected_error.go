package java

import (
	lib "github.com/nedpals/errgoengine"
	sitter "github.com/smacker/go-tree-sitter"
)

type characterExpectedFixKind int

const (
	characterExpectedFixUnknown      characterExpectedFixKind = 0
	characterExpectedFixWrapFunction characterExpectedFixKind = iota
)

type characterExpectedErrorCtx struct {
	fixKind characterExpectedFixKind
}

var CharacterExpectedError = lib.ErrorTemplate{
	Name:              "CharacterExpectedError",
	Pattern:           comptimeErrorPattern(`'(?P<character>\S+)' expected`),
	StackTracePattern: comptimeStackTracePattern,
	OnAnalyzeErrorFn: func(cd *lib.ContextData, m *lib.MainError) {
		iCtx := characterExpectedErrorCtx{}

		// TODO: check if node is parsable
		rootNode := m.Document.Tree.RootNode()
		cursor := sitter.NewTreeCursor(rootNode)
		rawNearestMissingNode := nearestMissingNodeFromPos(cursor, m.ErrorNode.StartPos)
		nearestMissingNode := lib.WrapNode(m.Document, rawNearestMissingNode)
		m.Nearest = nearestMissingNode
		m.Context = iCtx
	},
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		gen.Add("This error occurs when there is an unexpected character in the code, and '%s' is expected.", cd.Variables["character"])

		// ctx := cd.MainError.Context.(CharacterExpectedErrorCtx)

		// switch ctx.kind {
		// case cannotBeAppliedMismatchedArgCount:
		// 	gen.Add("This error occurs when there is an attempt to apply a method with an incorrect number of arguments.")
		// case cannotBeAppliedMismatchedArgType:
		// 	gen.Add("This error occurs when there is an attempt to apply a method with arguments that do not match the method signature.")
		// default:
		// 	gen.Add("unable to determine.")
		// }
	},
	OnGenBugFixFn: func(cd *lib.ContextData, gen *lib.BugFixGenerator) {
		// ctx := cd.MainError.Context.(characterExpectedErrorCtx)

		gen.Add("Add the missing character", func(s *lib.BugFixSuggestion) {
			s.AddStep("Ensure that the array declaration has the correct syntax by adding the missing `%s`.", cd.Variables["character"]).
				AddFix(lib.FixSuggestion{
					NewText:       cd.Variables["character"],
					StartPosition: cd.MainError.Nearest.StartPosition(),
					EndPosition:   cd.MainError.Nearest.StartPosition(),
				})
		})

		// switch ctx.fixKind {
		// case characterExpectedFixWrapFunction:
		// 	gen.Add("Use the correct syntax", func(s *lib.BugFixSuggestion) {
		// 		startPos := cd.MainError.Nearest.StartPosition()
		// 		space := getSpaceFromBeginning(cd.MainError.Document, startPos.Line, startPos.Column)

		// 		s.AddStep("Use a valid statement or expression within a method or block.").
		// 			AddFix(lib.FixSuggestion{
		// 				NewText: space + "public void someMethod() {\n" + space,
		// 				StartPosition: lib.Position{
		// 					Line: cd.MainError.Nearest.StartPosition().Line,
		// 				},
		// 				EndPosition: lib.Position{
		// 					Line: cd.MainError.Nearest.StartPosition().Line,
		// 				},
		// 			}).
		// 			AddFix(lib.FixSuggestion{
		// 				NewText:       "\n" + space + "}",
		// 				StartPosition: cd.MainError.Nearest.EndPosition(),
		// 				EndPosition:   cd.MainError.Nearest.EndPosition(),
		// 			})
		// 	})
		// }
	},
}
