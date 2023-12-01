package python

import (
	_ "embed"

	lib "github.com/nedpals/errgoengine"
	"github.com/smacker/go-tree-sitter/python"
)

//go:embed symbols.txt
var symbols string

var Language = &lib.Language{
	Name:              "Python",
	FilePatterns:      []string{".py"},
	SitterLanguage:    python.GetLanguage(),
	StackTracePattern: `\s+File "(?P<path>\S+)", line (?P<position>\d+), in (?P<symbol>\S+)`,
	ErrorPattern:      `Traceback \(most recent call last\):$stacktrace$message`,
	AnalyzerFactory: func(cd *lib.ContextData) lib.LanguageAnalyzer {
		return &pyAnalyzer{cd}
	},
	SymbolsToCapture: symbols,
}

type pyAnalyzer struct {
	*lib.ContextData
}

func (an *pyAnalyzer) FallbackSymbol() lib.Symbol {
	return lib.Builtin("any")
}

func (an *pyAnalyzer) AnalyzeNode(n lib.SyntaxNode) lib.Symbol {
	// TODO:
	return lib.Builtin("void")
}

func (an *pyAnalyzer) AnalyzeImport(params lib.ImportParams) lib.ResolvedImport {
	// TODO:

	return lib.ResolvedImport{
		Path: "",
	}
}
