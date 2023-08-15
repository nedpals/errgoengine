package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

type Position struct {
	Line   int
	Column int
	Index  int
}

type SymbolKind int

const (
	SymbolKindUnknown  SymbolKind = 0
	SymbolKindBuiltin  SymbolKind = iota
	SymbolKindClass    SymbolKind = iota
	SymbolKindFunction SymbolKind = iota
	SymbolKindVariable SymbolKind = iota
)

type Symbol struct {
	Name         string
	Kind         SymbolKind
	ReturnSymbol *Symbol
	ValueSymbol  *Symbol
	Children     *SymbolTree
	Location
}

func (sym *Symbol) Get(field string) *Symbol {
	if sym.Children != nil {
		for symName, sym := range sym.Children.Symbols {
			if symName == field {
				return sym
			}
		}
	}
	return nil
}

func BuiltinSymbol(name string) *Symbol {
	return &Symbol{
		Name: name,
		Kind: SymbolKindBuiltin,
	}
}

type SymbolTree struct {
	StartPos     Position
	EndPos       Position
	DocumentPath string
	Symbols      map[string]*Symbol
	Scopes       []*SymbolTree
}

func (tree *SymbolTree) Add(sym *Symbol) {
	if tree.Symbols == nil {
		tree.Symbols = make(map[string]*Symbol)
	}

	tree.Symbols[sym.Name] = sym
	// TODO: create tree both in the parent and in the child symbol

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

type Query string

type SymbolCapturer interface {
	Query() *sitter.Query
}

func countSuffix(str string, s byte) int {
	c := 0
	for i := len(str) - 1; i >= 0; i-- {
		if str[i] == s {
			c++
		} else {
			return c
		}
	}
	return 0
}

type SymbolCapture struct {
	Query    string
	Kind     SymbolKind
	Field    string
	Optional bool

	NameNode       *SymbolCapture
	ParameterNodes *SymbolCapture
	ReturnTypeNode *SymbolCapture // should be nil for top-level symbols
	ContentNode    *SymbolCapture
	BodyNode       *SymbolCapture

	Children []*SymbolCapture
}

func (cap *SymbolCapture) QueryString() string {
	sb := &strings.Builder{}
	cap.generateQueryString("sym", "", sb)
	return sb.String()
}

func (cap SymbolCapture) generateQueryString(prefix, tag string, sb *strings.Builder) {
	isAlternations := strings.HasPrefix(cap.Query, "[") && strings.HasSuffix(cap.Query, "]")
	isSingle := len(cap.Children) < 2
	parCount := countSuffix(cap.Query, ')')

	if len(cap.Field) != 0 {
		sb.WriteString(cap.Field)
		sb.WriteString(": ")
	}

	if !isAlternations {
		sb.WriteByte('(')
	}

	if parCount > 0 {
		sb.WriteString(cap.Query[:len(cap.Query)-parCount])
		sb.WriteByte(' ')
	} else {
		sb.WriteString(cap.Query)
	}

	if len(cap.Children) != 0 {
		if !isSingle {
			sb.WriteByte('\n')
			sb.WriteByte('[')
		}

		for i, c := range cap.Children {
			sb.WriteByte('\n')
			c.generateQueryString(
				fmt.Sprintf(
					"%s.child.%d",
					prefix,
					i,
				),
				"",
				sb,
			)
		}

		if !isSingle {
			sb.WriteString("\n]*")

			if len(tag) != 0 {
				sb.WriteString(" @" + prefix + "." + tag)
			}
		}
	} else {
		if cap.ReturnTypeNode != nil {
			sb.WriteRune('\n')
			cap.ReturnTypeNode.generateQueryString(prefix, "return-type", sb)
		}

		if cap.NameNode != nil {
			sb.WriteRune('\n')
			cap.NameNode.generateQueryString(prefix, "name", sb)
		}

		if cap.ParameterNodes != nil {
			sb.WriteRune('\n')
			cap.ParameterNodes.generateQueryString(prefix, "parameters", sb)
		}

		if cap.ContentNode != nil {
			sb.WriteRune('\n')
			cap.ContentNode.generateQueryString(prefix, "content", sb)
		}

		if cap.BodyNode != nil {
			sb.WriteRune('\n')
			cap.BodyNode.generateQueryString(prefix, "body", sb)
		}
	}

	if parCount > 0 {
		sb.WriteString(strings.Repeat(")", parCount))
	}

	if !isAlternations {
		sb.WriteByte(')')
	}

	if cap.Optional {
		sb.WriteByte('?')
	}

	if isSingle && len(tag) != 0 {
		sb.WriteString(" @")

		if len(prefix) != 0 {
			sb.WriteString(prefix)
			sb.WriteByte('.')
		}

		sb.WriteString(tag)
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

type MainError struct {
	ErrorNode *STGNode
	Document  *Document
	Nearest   Node
}

func (err MainError) DocumentPath() string {
	return err.ErrorNode.DocumentPath
}

// TODO: add import dependency graph for finding third-party symbols
type ContextData struct {
	WorkingPath     string
	Variables       map[string]string
	StackTraceGraph StackTraceGraph
	Documents       map[string]*Document
	Symbols         map[string]*SymbolTree
	MainError       MainError
}

func (data *ContextData) FindSymbol(name string) *Symbol {
	if data.MainError.Document == nil {
		return nil
	}

	// builtin symbols are already referenced inside the nodevalueanalyzer

	// TODO: improve this for later
	symbolsFromDoc := data.Symbols[data.MainError.DocumentPath()]
	for symName, sym := range symbolsFromDoc.Symbols {
		if symName == name {
			return sym
		}
	}

	return nil
}

func (data *ContextData) AnalyzeValue(n Node) *Symbol {
	if data.MainError.Document == nil || data.MainError.Document.Language.ValueAnalyzer == nil {
		return nil
	}

	return data.MainError.Document.
		Language.ValueAnalyzer(&NodeValueAnalyzer{data}, n)
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
}

var symPrefixRegex = regexp.MustCompile(`^sym.(\d+)`)

func (an *Analyzer) AnalyzeTree(tree *sitter.Tree) {
	rootNode := tree.RootNode()
	queryCursor := sitter.NewQueryCursor()
	sb := &strings.Builder{}

	sb.WriteString("[")
	for _, sc := range an.doc.Language.SymbolsToCapture {
		sc.generateQueryString(fmt.Sprintf("sym.%d", sc.Kind), "", sb)
	}
	sb.WriteString("]+ @sym")

	rawQuery := sb.String()
	q, err := sitter.NewQuery([]byte(rawQuery), an.doc.Language.SitterLanguage)
	if err != nil {
		panic(err)
	}

	queryCursor.Exec(q, rootNode)

	for {
		m, ok := queryCursor.NextMatch()
		if !ok {
			break
		} else if len(m.Captures) == 0 {
			continue
		}

		// group first the information
		captured := map[string]Node{}
		firstMatchCname := ""
		for _, c := range m.Captures {
			key := q.CaptureNameForId(c.Index)
			captured[key] = wrapNode(an.doc, c.Node)
			if len(firstMatchCname) == 0 {
				if matches := symPrefixRegex.FindStringSubmatch(key); len(matches) != 0 {
					firstMatchCname = matches[1]
				}
			}
		}

		if len(captured) == 0 {
			continue
		}

		identifiedKind := SymbolKindUnknown
		convertedKind, _ := strconv.Atoi(firstMatchCname)
		identifiedKind = SymbolKind(convertedKind)

		// rename map entries
		for k := range captured {
			renamed := strings.TrimPrefix(k, fmt.Sprintf("sym.%d.", identifiedKind))
			if renamed == k {
				continue
			}

			captured[renamed] = captured[k]
			delete(captured, k)
		}

		// each item contains
		// - node
		// - content
		// - position
		// - item name (sym.children.0.name for example)
		// TODO: children

		body := captured["body"]
		an.contextData.AddSymbol(&Symbol{
			Name: captured["name"].Text(),
			Kind: identifiedKind,
			// TODO ValueSymbol: ,
			// ReturnSymbol: an.parent.Children.Symbols[captured["return-type"].Text()], // TODO
			Location: captured["sym"].Location(),
			Children: &SymbolTree{
				StartPos:     body.StartPosition(),
				EndPos:       body.EndPosition(),
				DocumentPath: an.doc.Path,
				Symbols:      map[string]*Symbol{},
			},
		})
	}
}

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

func (n Node) Child(idx int) Node {
	cNode := n.Node.Child(idx)
	return wrapNode(n.doc, cNode)
}

func (n Node) StartPosition() Position {
	p := n.Node.StartPoint()
	return Position{
		Line:   int(p.Row),
		Column: int(p.Column),
		Index:  int(n.Node.StartByte()),
	}
}

func (n Node) EndPosition() Position {
	p := n.Node.EndPoint()
	return Position{
		Line:   int(p.Row),
		Column: int(p.Column),
		Index:  int(n.Node.StartByte()),
	}
}

func (n Node) Location() Location {
	return Location{
		DocumentPath: n.doc.Path,
		Position:     n.StartPosition(),
	}
}

func wrapNode(doc *Document, n *sitter.Node) Node {
	return Node{
		text: n.Content([]byte(doc.Contents)),
		doc:  doc,
		Node: n,
	}
}

type ValueNodeTransformer func(ValueNodeTransformer, *sitter.Node) *sitter.Node

type NodeValueAnalyzerFn func(*NodeValueAnalyzer, Node) *Symbol

type NodeValueAnalyzer struct {
	context *ContextData
}

func (an *NodeValueAnalyzer) Analyze(node Node) *Symbol {
	return an.context.AnalyzeValue(node)
}

func (an *NodeValueAnalyzer) Find(name string) *Symbol {
	// Find local symbols first

	// TODO: find nearest symbol tree
	// path := an.context.MainError.DocumentPath()
	// tree := an.context.Symbols[path]
	// for symName, sym := range

	return an.context.FindSymbol(name)
}

type Language struct {
	isCompiled           bool
	Name                 string
	FilePatterns         []string
	SitterLanguage       *sitter.Language
	StackTracePattern    *regexp.Regexp
	BuiltinTypes         []string
	BuiltinSymbols       []*Symbol
	SymbolsToCapture     []SymbolCapture
	LocationConverter    func(path, pos string) Location
	ValueNodeTransformer ValueNodeTransformer
	ValueAnalyzer        NodeValueAnalyzerFn
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

func (lang *Language) AddTemplate(template ErrorTemplate) *CompiledErrorTemplate {
	return errorTemplates.Add(lang, template)
}

type ErrorTemplate struct {
	Name              string
	Pattern           string
	StackTracePattern string
	OnGenExplainFn    func(*ContextData) string
	OnGenBugFixFn     func(*ContextData) []BugFix
}

type CompiledErrorTemplate struct {
	ErrorTemplate
	Language *Language
	Pattern  *regexp.Regexp
}

type ErrorTemplates []*CompiledErrorTemplate

func (tmps *ErrorTemplates) Add(language *Language, template ErrorTemplate) *CompiledErrorTemplate {
	patternForCompile := "(?m)^" + template.Pattern + `(?P<stacktrace>(?:.|\s)*)$`
	compiledPattern, err := regexp.Compile(patternForCompile)
	if err != nil {
		// TODO: should not panic!
		panic(err)
	}

	*tmps = append(*tmps, &CompiledErrorTemplate{
		ErrorTemplate: template,
		Language:      language,
		Pattern:       compiledPattern,
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
			if len(groupNames[idx]) == 0 {
				continue
			}

			contextData.AddVariable(groupNames[idx], matchedContent)
		}
	}

	// extract stack trace
	rawStackTraces := contextData.Variables["stacktrace"]
	symbolGroupIdx := template.Language.StackTracePattern.SubexpIndex("symbol")
	pathGroupIdx := template.Language.StackTracePattern.SubexpIndex("path")
	posGroupIdx := template.Language.StackTracePattern.SubexpIndex("position")

	for _, rawStackTrace := range strings.Split(rawStackTraces, "\n") {
		submatches := template.Language.StackTracePattern.FindAllStringSubmatch(rawStackTrace, -1)
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

		stLoc := template.Language.LocationConverter(rawPath, rawPos)
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
		selectedLanguage := template.Language
		if !selectedLanguage.MatchPath(node.DocumentPath) {
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

		analyzer.AnalyzeTree(tree)
	}

	// error translation
	return TranslateError(template.ErrorTemplate, contextData)
}

func locateNearestNode(cursor *sitter.TreeCursor, pos Position) *sitter.Node {
	cursor.GoToFirstChild()
	defer cursor.GoToParent()

	for {
		currentNode := cursor.CurrentNode()
		pointA := currentNode.StartPoint()
		pointB := currentNode.EndPoint()

		if pointA.Row+1 == uint32(pos.Line) {
			return currentNode
		} else if uint32(pos.Line) >= pointA.Row+1 && uint32(pos.Line) <= pointB.Row+1 {
			return locateNearestNode(cursor, pos)
		} else if !cursor.GoToNextSibling() {
			return nil
		}
	}
}

func TranslateError(template ErrorTemplate, contextData *ContextData) string {
	// locate main error
	for _, node := range contextData.StackTraceGraph {
		if !strings.HasPrefix(node.DocumentPath, contextData.WorkingPath) {
			continue
		}

		// get nearest node
		doc := contextData.Documents[node.DocumentPath]
		nearest := doc.Tree.RootNode().NamedDescendantForPointRange(
			sitter.Point{Row: uint32(node.Line)},
			sitter.Point{Row: uint32(node.Line)},
		)

		if nearest.StartPoint().Row != uint32(node.Line) {
			cursor := sitter.NewTreeCursor(nearest)
			nearest = locateNearestNode(cursor, node.Position)
		}

		if doc.Language.ValueNodeTransformer != nil {
			nearest = doc.Language.ValueNodeTransformer(
				doc.Language.ValueNodeTransformer,
				nearest,
			)
		}

		contextData.MainError = struct {
			ErrorNode *STGNode
			Document  *Document
			Nearest   Node
		}{
			ErrorNode: &node,
			Document:  doc,
			Nearest:   wrapNode(doc, nearest),
		}

		break
	}

	// execute error generator function
	explanation := template.OnGenExplainFn(contextData)

	return explanation
}

func main() {
	JavaLanguage.AddTemplate(NullPointerException)
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
