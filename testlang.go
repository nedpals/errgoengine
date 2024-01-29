package errgoengine

import (
	"context"
)

var TestLanguage = &Language{
	Name:              "TestLang",
	FilePatterns:      []string{".test"},
	StackTracePattern: `\sin (?P<symbol>\S+) at (?P<path>\S+):(?P<position>\d+)`,
	LocationConverter: func(ctx LocationConverterContext) Location {
		return Location{
			DocumentPath: ctx.Path,
			StartPos:     Position{Line: 0, Column: 0, Index: 0},
			EndPos:       Position{Line: 0, Column: 0, Index: 0},
		}
	},
	AnalyzerFactory: func(cd *ContextData) LanguageAnalyzer {
		return &testAnalyzer{}
	},
	SymbolsToCapture: `
(expression_statement
	(assignment
		left: (identifier) @assignment.name
		right: (identifier) @assignment.content) @assignment)
	`,
}

type testAnalyzer struct {
	*ContextData
}

func (an *testAnalyzer) FallbackSymbol() Symbol {
	return Builtin("any")
}

func (an *testAnalyzer) FindSymbol(name string) Symbol {
	return nil
}

func (an *testAnalyzer) AnalyzeNode(_ context.Context, n SyntaxNode) Symbol {
	// TODO:
	return Builtin("void")
}

func (an *testAnalyzer) AnalyzeImport(params ImportParams) ResolvedImport {
	// TODO:

	return ResolvedImport{
		Path: "",
	}
}
