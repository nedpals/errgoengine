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
	FallbackSymbol() Symbol
	AnalyzeNode(SyntaxNode) Symbol
	AnalyzeImport(ImportParams) ResolvedImport
}

type Language struct {
	isCompiled        bool
	stackTraceRegex   *regexp.Regexp
	stubFs            *StubFS
	Name              string
	FilePatterns      []string
	SitterLanguage    *sitter.Language
	StackTracePattern string
	ErrorPattern      string
	SymbolsToCapture  string
	LocationConverter func(path, pos string) Location
	AnalyzerFactory   func(cd *ContextData) LanguageAnalyzer
	OnGenStubFS       func(fs *StubFS)
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

	if lang.stubFs == nil {
		lang.stubFs = &StubFS{}
		lang.OnGenStubFS(lang.stubFs)
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
		StartPos:     Position{Line: trueLine},
		EndPos:       Position{Line: trueLine},
	}
}
