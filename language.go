package errgoengine

import (
	"fmt"
	"regexp"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

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
	ValueAnalyzer     func(NodeValueAnalyzer, Node) Symbol
	ImportResolver    func(NodeValueAnalyzer, ImportParams) ResolvedImport
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

	if lang.ImportResolver == nil {
		panic(fmt.Sprintf("[Language -> %s] ImportResolver must not be nil", lang.Name))
	}

	if lang.SymbolsToCapture != nil && lang.ValueAnalyzer == nil {
		panic(fmt.Sprintf("[Language -> %s] ValueAnalyzer must not be nil", lang.Name))
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
