package lib

import (
	"regexp"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

type Language struct {
	isCompiled        bool
	Name              string
	FilePatterns      []string
	SitterLanguage    *sitter.Language
	StackTracePattern *regexp.Regexp
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

	// TODO: should this be removed or not? hmmmmm

	lang.isCompiled = true
}
