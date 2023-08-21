package lib

import (
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

	lang.stackTraceRegex = regexp.MustCompile("(?m)" + lang.StackTracePattern)
	lang.isCompiled = true
}
