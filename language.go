package errgoengine

import (
	"fmt"
	"regexp"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

type NodeValueAnalyzer interface {
	FindSymbol(name string, pos int) Symbol
	AnalyzeValue(n SyntaxNode) Symbol
}

type LanguageAnalyzer interface {
	AnalyzeNode(SyntaxNode) Symbol
	AnalyzeImport(ImportParams) ResolvedImport
}

type Language struct {
	isCompiled        bool
	stackTraceRegex   *regexp.Regexp
	Name              string
	FilePatterns      []string
	SitterLanguage    *sitter.Language
	StackTracePattern string
	ErrorPattern      string
	SymbolsToCapture  ISymbolCaptureList
	LocationConverter func(path, pos string) Location
	AnalyzerFactory   func(cd *ContextData) LanguageAnalyzer
}

func (lang *Language) MatchPath(path string) bool {
	for _, ext := range lang.FilePatterns {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	return false
}

func (lang *Language) Compile() {
	if lang.isCompiled {
		return
	}

	if lang.LocationConverter == nil {
		lang.LocationConverter = DefaultLocationConverter
	}

	if len(lang.StackTracePattern) != 0 {
		lang.stackTraceRegex = regexp.MustCompile("(?m)" + lang.StackTracePattern)
	}

	if lang.AnalyzerFactory == nil {
		panic(fmt.Sprintf("[Language -> %s] AnalyzerFactory must not be nil", lang.Name))
	}

	lang.isCompiled = true
}

// SetTemplateStackTraceRegex sets the language's regex pattern directly. for testing purposes only
func SetTemplateStackTraceRegex(lang *Language, pattern *regexp.Regexp) {
	lang.stackTraceRegex = pattern
}

func DefaultLocationConverter(path, pos string) Location {
	var trueLine int
	if _, err := fmt.Sscanf(pos, "%d", &trueLine); err != nil {
		panic(err)
	}
	return Location{
		DocumentPath: path,
		Position:     Position{Line: trueLine},
	}
}
