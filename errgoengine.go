package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

type Position struct {
	Line   int
	Column int
	Index  int
}

type SymbolType int

const (
	SymbolTypeUnknown  SymbolType = 0
	SymbolTypeBuiltin  SymbolType = iota
	SymbolTypeClass    SymbolType = iota
	SymbolTypeFunction SymbolType = iota
	SymbolTypeVariable SymbolType = iota
)

type Symbol struct {
	Name         string
	Type         SymbolType
	ReturnSymbol *Symbol
	Children     *SymbolTree
	Location
}

type SymbolTree struct {
	StartPos     Position
	EndPos       Position
	DocumentPath string
	Symbols      map[string]*Symbol
}

func (tree *SymbolTree) Add(sym *Symbol) {
	if tree.Symbols == nil {
		tree.Symbols = make(map[string]*Symbol)
	}

	tree.Symbols[sym.Name] = sym

	if sym.Position.Index < tree.StartPos.Index {
		tree.StartPos = Position{
			Line:   sym.Line,
			Column: sym.Column,
			Index:  sym.Index,
		}
	}

	if sym.Position.Index > tree.EndPos.Index {
		tree.EndPos = Position{
			Line:   sym.Line,
			Column: sym.Column,
			Index:  sym.Index,
		}
	}
}

type Location struct {
	DocumentPath string
	Position
}

type Document struct {
	Path     string
	Contents string
	Language *Language
	Tree     *sitter.Tree
}

type ContextData struct {
	WorkingPath     string
	Variables       map[string]string
	StackTraceGraph StackTraceGraph
	Documents       map[string]*Document
	Symbols         map[string]*SymbolTree
}

func (data *ContextData) LocateMainDoc() *Document {
	for _, node := range data.StackTraceGraph {
		if node.IsMain {
			return data.Documents[node.DocumentPath]
		}
	}
	return nil
}

func (data *ContextData) AddVariable(name string, value string) {
	if data.Variables == nil {
		data.Variables = make(map[string]string)
	}

	data.Variables[name] = value
}

func (data *ContextData) AddDocument(path, contents string, lang *Language, tree *sitter.Tree) *Document {
	if data.Documents == nil {
		data.Documents = make(map[string]*Document)
	}

	doc := &Document{
		Path:     path,
		Language: lang,
		Contents: contents,
		Tree:     tree,
	}

	data.Documents[path] = doc
	return doc
}

func (data *ContextData) AddSymbol(sym *Symbol) *Symbol {
	if data.Symbols == nil {
		data.Symbols = make(map[string]*SymbolTree)
	}

	if data.Symbols[sym.Location.DocumentPath] == nil {
		data.Symbols[sym.Location.DocumentPath] = &SymbolTree{
			DocumentPath: sym.Location.DocumentPath,
			Symbols:      make(map[string]*Symbol),
		}
	}

	data.Symbols[sym.Location.DocumentPath].Add(sym)
	return sym
}

type BugFix struct {
	Content string // explanation
	Code    string
}

type Analyzer struct {
	contextData *ContextData
	doc         *Document
	parent      *Symbol
}

func (an *Analyzer) SetParent(newParent *Symbol) {
	an.parent = newParent
	// if newParent != nil {
	// 	fmt.Println("=== set parent to " + newParent.Name)
	// }
}

func (an *Analyzer) AnalyzeNode(node *sitter.Node) {
	nodeType := node.Type()
	if extractor, ok := an.doc.Language.SymbolExtractors[nodeType]; ok {
		oldParent := an.parent
		wrappedNode := wrapNode(an.doc, node)
		extractor(wrappedNode, an)
		an.SetParent(oldParent)
	} else {
		cursor := sitter.NewTreeCursor(node)
		defer cursor.Close()

		an.AnalyzeCursor(cursor)
	}
}

func (an *Analyzer) AddSymbol(sym *Symbol) *Symbol {
	if an.parent != nil {
		if an.parent.Children == nil {
			an.parent.Children = &SymbolTree{}
		}

		an.parent.Children.Add(sym)
	} else {
		an.contextData.AddSymbol(sym)
	}

	return sym
}

func (an *Analyzer) AddSymbolFromWrappedNode(node Node, typ SymbolType, loc Location) *Symbol {
	return an.AddSymbol(&Symbol{
		Name:     node.Text(),
		Type:     typ,
		Location: loc,
	})
}

func (an *Analyzer) AnalyzeCursor(cursor *sitter.TreeCursor) {
	if !cursor.GoToFirstChild() {
		return
	}

	for {
		node := cursor.CurrentNode()
		an.AnalyzeNode(node)

		if !cursor.GoToNextSibling() {
			break
		}
	}
}

type SymbolExtractorFn func(Node, *Analyzer)

type Node struct {
	*sitter.Node
	doc  *Document
	text string
}

func (n Node) Text() string {
	return n.text
}

func (n Node) ChildByFieldName(field string) Node {
	cNode := n.Node.ChildByFieldName(field)
	return wrapNode(n.doc, cNode)
}

func (n Node) Location() Location {
	pointA := n.Node.StartPoint()
	return Location{
		DocumentPath: n.doc.Path,
		Position: Position{
			Line:   int(pointA.Row),
			Column: int(pointA.Column),
			Index:  int(n.Node.StartByte()),
		},
	}
}

func wrapNode(doc *Document, n *sitter.Node) Node {
	return Node{
		text: n.Content([]byte(doc.Contents)),
		doc:  doc,
		Node: n,
	}
}

type Language struct {
	Name             string
	FilePatterns     []string
	SitterLanguage   *sitter.Language
	IsCompiled       bool
	BuiltinTypes     []string
	BuiltinSymbols   []*Symbol
	SymbolExtractors map[string]SymbolExtractorFn
}

func (lang *Language) Compile() {
	if lang.IsCompiled {
		return
	}

	if len(lang.BuiltinTypes) != 0 && lang.BuiltinSymbols == nil {
		lang.BuiltinSymbols = make([]*Symbol, 0, len(lang.BuiltinTypes))
	}

	for _, typ := range lang.BuiltinTypes {
		lang.BuiltinSymbols = append(lang.BuiltinSymbols, &Symbol{
			Name: typ,
			Type: SymbolTypeBuiltin,
		})
	}

	lang.IsCompiled = true
}

type ErrorTemplate struct {
	Name                string
	Pattern             string
	StackTracePattern   string
	LocationConverterFn func(string, string) Location
	OnGenExplainFn      func(*Document, *ContextData) string
	OnGenBugFixFn       func(*Document, *ContextData) []BugFix
}

type CompiledErrorTemplate struct {
	ErrorTemplate
	Pattern           *regexp.Regexp
	StackTracePattern *regexp.Regexp
}

type ErrorTemplates []*CompiledErrorTemplate

func (tmps *ErrorTemplates) Add(template ErrorTemplate) *CompiledErrorTemplate {
	patternForCompile := "(?m)^" + template.Pattern + `(?P<stacktrace>(?:.|\s)*)$`
	compiledPattern, err := regexp.Compile(patternForCompile)
	if err != nil {
		// TODO: should not panic!
		panic(err)
	}

	compiledStacktracePattern, err := regexp.Compile(template.StackTracePattern)
	if err != nil {
		panic(err)
	}

	*tmps = append(*tmps, &CompiledErrorTemplate{
		ErrorTemplate:     template,
		Pattern:           compiledPattern,
		StackTracePattern: compiledStacktracePattern,
	})

	return (*tmps)[len(*tmps)-1]
}

func (tmps ErrorTemplates) Find(msg string) *CompiledErrorTemplate {
	for _, tmp := range tmps {
		if tmp.Pattern.MatchString(msg) {
			return tmp
		}
	}
	return nil
}

var errorTemplates = ErrorTemplates{}
var languages = []*Language{JavaLanguage}

func Analyze(workingPath string, msg string) string {
	template := errorTemplates.Find(msg)
	if template == nil {
		panic("Template not found!")
	}

	// initial context data extraction
	contextData := &ContextData{WorkingPath: workingPath}
	groupNames := template.Pattern.SubexpNames()
	for _, submatches := range template.Pattern.FindAllStringSubmatch(msg, -1) {
		for idx, matchedContent := range submatches {
			// fmt.Printf("idx = %d, matchedContent = %v\n", idx, matchedContent)
			if len(groupNames[idx]) == 0 {
				continue
			}

			contextData.AddVariable(groupNames[idx], matchedContent)
		}
	}

	// extract stack trace
	rawStackTraces := contextData.Variables["stacktrace"]
	symbolGroupIdx := template.StackTracePattern.SubexpIndex("symbol")
	pathGroupIdx := template.StackTracePattern.SubexpIndex("path")
	posGroupIdx := template.StackTracePattern.SubexpIndex("position")

	for _, rawStackTrace := range strings.Split(rawStackTraces, "\n") {
		submatches := template.StackTracePattern.FindAllStringSubmatch(rawStackTrace, -1)
		if len(submatches) == 0 {
			continue
		}

		matches := submatches[0]
		rawSymbolName := matches[symbolGroupIdx]
		rawPath := matches[pathGroupIdx]
		rawPos := matches[posGroupIdx]

		// convert relative paths to absolute for parsing
		if !filepath.IsAbs(rawPath) {
			rawPath = filepath.Clean(filepath.Join(workingPath, rawPath))
		}

		stLoc := template.LocationConverterFn(rawPath, rawPos)
		if contextData.StackTraceGraph == nil {
			contextData.StackTraceGraph = StackTraceGraph{}
		}

		contextData.StackTraceGraph.Add(rawSymbolName, stLoc)
	}

	// open contents of the extracted stack file locations
	parser := sitter.NewParser()

	for _, node := range contextData.StackTraceGraph {
		contents, err := os.ReadFile(node.DocumentPath)
		if err != nil {
			panic(err)
		}

		// check matched languages
		var selectedLanguage *Language

		for _, lang := range languages {
			for _, ext := range lang.FilePatterns {
				if strings.HasSuffix(node.DocumentPath, ext) {
					selectedLanguage = lang
					break
				}
			}

			if selectedLanguage != nil {
				break
			}
		}

		if selectedLanguage == nil {
			panic("No language found for " + node.DocumentPath)
		}

		// compile builtin types first of a language if not available
		selectedLanguage.Compile()

		// do semantic analysis
		parser.SetLanguage(selectedLanguage.SitterLanguage)
		tree, err := parser.ParseCtx(context.Background(), nil, contents)
		if err != nil {
			panic(err)
		}

		doc := contextData.AddDocument(node.DocumentPath, string(contents), selectedLanguage, tree)
		parser.Reset()

		analyzer := &Analyzer{
			contextData: contextData,
			doc:         doc,
		}

		analyzer.AnalyzeNode(tree.RootNode())
	}

	// error translation
	return TranslateError(template.ErrorTemplate, contextData)
}

func TranslateError(template ErrorTemplate, contextData *ContextData) string {
	// locate main culprit file
	doc := contextData.LocateMainDoc()

	// execute error generator function
	explanation := template.OnGenExplainFn(doc, contextData)

	return explanation
}

func main() {
	errorTemplates.Add(NullPointerException)
	wd, _ := os.Getwd()

	var errMsg string

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		if len(errMsg) != 0 {
			errMsg += "\n"
		}

		errMsg += scanner.Text()
	}

	fmt.Println(Analyze(wd, errMsg))
}
