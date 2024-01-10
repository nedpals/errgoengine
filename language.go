package errgoengine

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
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
	FindSymbol(name string) Symbol
	AnalyzeNode(context.Context, SyntaxNode) Symbol
	AnalyzeImport(ImportParams) ResolvedImport
}

type Language struct {
	isCompiled        bool
	stackTraceRegex   *regexp.Regexp
	externSymbols     map[string]*SymbolTree
	Name              string
	FilePatterns      []string
	SitterLanguage    *sitter.Language
	StackTracePattern string
	ErrorPattern      string
	SymbolsToCapture  string
	LocationConverter func(path, pos string) Location
	AnalyzerFactory   func(cd *ContextData) LanguageAnalyzer
	ExternFS          fs.ReadFileFS
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

	if err := lang.compileExternSymbols(); err != nil {
		panic(err)
	}

	lang.isCompiled = true
}

func (lang *Language) compileExternSymbols() error {
	if lang.isCompiled || lang.ExternFS == nil {
		return nil
	}

	lang.externSymbols = make(map[string]*SymbolTree)

	matches, err := fs.Glob(lang.ExternFS, "**/*.json")
	if err != nil {
		return err
	}

	for _, match := range matches {
		if err := lang.compileExternSymbol(match); err != nil {
			return err
		}
	}
}

func (lang *Language) compileExternSymbol(path string) error {
	if lang.isCompiled || lang.ExternFS == nil {
		return nil
	}

	file, err := lang.ExternFS.Open(path)
	if err != nil {
		return err
	}

	defer file.Close()

	var symTree SymbolTree
	if err := json.NewDecoder(file).Decode(&symTree); err != nil {
		return err
	}

	lang.externSymbols[path] = &symTree
	return nil
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
