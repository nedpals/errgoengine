package java

import (
	"context"

	lib "github.com/nedpals/errgoengine"
	sitter "github.com/smacker/go-tree-sitter"
)

type identifierExpectedFixKind int

const (
	identifierExpectedFixUnknown      identifierExpectedFixKind = 0
	identifierExpectedFixWrapFunction identifierExpectedFixKind = iota
)

type identifierExpectedErrorCtx struct {
	fixKind identifierExpectedFixKind
}

var IdentifierExpectedError = lib.ErrorTemplate{
	Name:              "IdentifierExpectedError",
	Pattern:           comptimeErrorPattern(`<identifier> expected`),
	StackTracePattern: comptimeStackTracePattern,
	OnAnalyzeErrorFn: func(cd *lib.ContextData, m *lib.MainError) {
		iCtx := identifierExpectedErrorCtx{}

		// TODO: check if node is parsable
		if tree, err := sitter.ParseCtx(
			context.Background(),
			[]byte(m.Nearest.Text()),
			m.Document.Language.SitterLanguage,
		); err == nil && !tree.IsError() {
			iCtx.fixKind = identifierExpectedFixWrapFunction
		}

		m.Context = iCtx
	},
	OnGenExplainFn: func(cd *lib.ContextData, gen *lib.ExplainGenerator) {
		gen.Add("This error occurs when an identifier is expected, but an expression is found in a location where a statement or declaration is expected.")

		// ctx := cd.MainError.Context.(IdentifierExpectedErrorCtx)

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
		ctx := cd.MainError.Context.(identifierExpectedErrorCtx)

		switch ctx.fixKind {
		case identifierExpectedFixWrapFunction:
			gen.Add("Use the correct syntax", func(s *lib.BugFixSuggestion) {
				startPos := cd.MainError.Nearest.StartPosition()
				space := getSpaceFromBeginning(cd.MainError.Document, startPos.Line, startPos.Column)

				s.AddStep("Use a valid statement or expression within a method or block.").
					AddFix(lib.FixSuggestion{
						NewText: space + "public void someMethod() {\n" + space,
						StartPosition: lib.Position{
							Line: cd.MainError.Nearest.StartPosition().Line,
						},
						EndPosition: lib.Position{
							Line: cd.MainError.Nearest.StartPosition().Line,
						},
					}).
					AddFix(lib.FixSuggestion{
						NewText:       "\n" + space + "}",
						StartPosition: cd.MainError.Nearest.EndPosition(),
						EndPosition:   cd.MainError.Nearest.EndPosition(),
					})
			})
		}
	},
}
