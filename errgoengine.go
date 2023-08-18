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

type SymbolTree struct {
	Parent       *SymbolTree
	StartPos     Position
	EndPos       Position
	DocumentPath string
	Symbols      map[string]Symbol
	Scopes       []*SymbolTree
}

func (tree *SymbolTree) CreateChildFromNode(n Node) *SymbolTree {
	return &SymbolTree{
		Parent:       tree,
		StartPos:     n.StartPosition(),
		EndPos:       n.EndPosition(),
		DocumentPath: tree.DocumentPath,
		Symbols:      map[string]Symbol{},
	}
}

func (tree *SymbolTree) Find(name string) Symbol {
	for _, sym := range tree.Symbols {
		if sym.Name() == name {
			return sym
		}
	}
	return nil
}

func (tree *SymbolTree) Add(sym Symbol) {
	if tree.Symbols == nil {
		tree.Symbols = make(map[string]Symbol)
	}

	tree.Symbols[sym.Name()] = sym
	// TODO: create tree both in the parent and in the child symbol

	if sym.Location().Position.Index < tree.StartPos.Index {
		tree.StartPos = sym.Location().Position
	}

	if sym.Location().Index > tree.EndPos.Index {
		tree.EndPos = sym.Location().Position
	}

	if cSym := CastChildrenSymbol(sym); cSym != nil {
		tree.Scopes = append(tree.Scopes, cSym.Children())
		cSym.Children().Parent = tree
	}
}

type Query string

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

func (cap SymbolCapture) SymKind() SymbolKind {
	return cap.Kind
}

func (cap SymbolCapture) Compile(prefix, tag string, sb *strings.Builder) {
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
			c.Compile(
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
			cap.ReturnTypeNode.Compile(prefix, "return-type", sb)
		}

		if cap.NameNode != nil {
			sb.WriteRune('\n')
			cap.NameNode.Compile(prefix, "name", sb)
		}

		if cap.ParameterNodes != nil {
			sb.WriteRune('\n')
			cap.ParameterNodes.Compile(prefix, "parameters", sb)
		}

		if cap.ContentNode != nil {
			sb.WriteRune('\n')
			cap.ContentNode.Compile(prefix, "content", sb)
		}

		if cap.BodyNode != nil {
			sb.WriteByte('\n')
			sb.WriteString("body: (_) @")
			if len(prefix) != 0 {
				sb.WriteString(prefix)
				sb.WriteByte('.')
			}
			sb.WriteString("body")
			// cap.BodyNode.Compile(prefix, "body", sb)
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
	WorkingPath         string
	CurrentDocumentPath string
	Variables           map[string]string
	StackTraceGraph     StackTraceGraph
	Documents           map[string]*Document
	Symbols             map[string]*SymbolTree
	MainError           MainError
}

func (data *ContextData) MainDocumentPath() string {
	if data.MainError.ErrorNode != nil {
		return data.MainError.DocumentPath()
	}
	return data.CurrentDocumentPath
}

func (data *ContextData) FindSymbol(name string) Symbol {
	if data.MainError.Document == nil {
		return nil
	}

	// TODO: improve this for later
	tree := data.Symbols[data.MainDocumentPath()]
	return tree.Find(name)
}

func (data *ContextData) AnalyzeValue(n Node) Symbol {
	return n.doc.Language.ValueAnalyzer(&NodeValueAnalyzer{data}, n)
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

func (data *ContextData) InitOrGetSymbolTree(docPath string) *SymbolTree {
	if data.Symbols == nil {
		data.Symbols = make(map[string]*SymbolTree)
	}

	if data.Symbols[docPath] == nil {
		data.Symbols[docPath] = &SymbolTree{
			DocumentPath: docPath,
			Symbols:      make(map[string]Symbol),
		}
	}

	return data.Symbols[docPath]
}

type BugFix struct {
	Content string // explanation
	Code    string
}

type Analyzer struct {
	contextData *ContextData
	doc         *Document
}

var symPrefix = "%d.sym.%d"

// ^\d+.sym.\d+
var symPrefixRegex = regexp.MustCompile(fmt.Sprintf("^%s", strings.ReplaceAll(symPrefix, "%d", `\d+`)))

type ISymbolCapture interface {
	Compile(prefix, tag string, sb *strings.Builder)
	SymKind() SymbolKind
}

type ISymbolCaptureList []ISymbolCapture

func (list ISymbolCaptureList) Compile(prefix, tag string, sb *strings.Builder) {
	sb.WriteString("[")
	for idx, sc := range list {
		sc.Compile(fmt.Sprintf(symPrefix, idx, sc.SymKind()), "", sb)
	}
	sb.WriteString("]+")
	if len(tag) != 0 {
		sb.WriteString(" @")
		if len(prefix) != 0 {
			sb.WriteString(prefix)
			sb.WriteByte('.')
		}
		sb.WriteString(tag)
	}
}

func (list ISymbolCaptureList) SymKind() SymbolKind {
	return SymbolKindUnknown
}

func (an *Analyzer) captureAndAnalyze(parent *SymbolTree, rootNode *sitter.Node, symbolCaptures ...ISymbolCapture) {
	if len(symbolCaptures) == 0 {
		return
	}

	if parent == nil {
		panic("Parent is null")
	}

	sb := &strings.Builder{}
	ISymbolCaptureList(symbolCaptures).Compile("", "sym", sb)
	q, err := sitter.NewQuery([]byte(sb.String()), an.doc.Language.SitterLanguage)
	if err != nil {
		panic(err)
	}

	queryCursor := sitter.NewQueryCursor()
	defer queryCursor.Close()

	queryCursor.Exec(q, rootNode)

	for i := 0; ; i++ {
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
			if len(firstMatchCname) == 0 && symPrefixRegex.MatchString(key) {
				firstMatchCname = key
			}
		}

		if len(captured) == 0 {
			continue
		}

		var identifiedKind SymbolKind
		var captureIdx int

		_, err := fmt.Sscanf(firstMatchCname, symPrefix, &captureIdx, &identifiedKind)
		if err != nil {
			panic(err)
		}

		// rename map entries
		for k := range captured {
			renamed := strings.TrimPrefix(k, fmt.Sprintf(symPrefix+".", captureIdx, identifiedKind))
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
		if body, ok := captured["body"]; ok {
			// returnSym = an.contextData.AnalyzeValue(body)
			childTree := parent.CreateChildFromNode(body)

			children := make(ISymbolCaptureList, 0)
			symCapture := symbolCaptures[captureIdx]

			switch any(symCapture).(type) {
			case SymbolCapture:
				children = symCaptureToListPtr(symCapture.(SymbolCapture).BodyNode.Children)
			case *SymbolCapture:
				children = symCaptureToListPtr(symCapture.(*SymbolCapture).BodyNode.Children)
			}

			an.captureAndAnalyze(childTree, body.Node, children...)
			parent.Add(&TopLevelSymbol{
				name:     captured["name"].Text(),
				kind:     identifiedKind,
				location: captured["sym"].Location(),
				children: childTree,
			})
		} else if content, ok := captured["content"]; ok {
			returnSym := an.contextData.AnalyzeValue(content)
			parent.Add(&VariableSymbol{
				name:       captured["name"].Text(),
				location:   captured["sym"].Location(),
				returnType: returnSym,
			})
		}
	}
}

func symCaptureToListPtr(list []*SymbolCapture) ISymbolCaptureList {
	captures := make(ISymbolCaptureList, len(list))
	for i, sc := range list {
		captures[i] = sc
	}
	return captures
}

func symCaptureToList(list []SymbolCapture) ISymbolCaptureList {
	captures := make(ISymbolCaptureList, len(list))
	for i, sc := range list {
		captures[i] = sc
	}
	return captures
}

func (an *Analyzer) AnalyzeTree(tree *sitter.Tree) {
	rootNode := tree.RootNode()
	captures := symCaptureToList(an.doc.Language.SymbolsToCapture)
	symTree := an.contextData.InitOrGetSymbolTree(an.doc.Path)
	an.contextData.CurrentDocumentPath = an.doc.Path
	an.captureAndAnalyze(symTree, rootNode, captures...)
	an.contextData.CurrentDocumentPath = ""
}

type ValueNodeTransformer func(ValueNodeTransformer, *sitter.Node) *sitter.Node

type NodeValueAnalyzerFn func(*NodeValueAnalyzer, Node) Symbol

type NodeValueAnalyzer struct {
	context *ContextData
}

func (an *NodeValueAnalyzer) Analyze(node Node) Symbol {
	return node.doc.Language.ValueAnalyzer(an, node)
}

func (an *NodeValueAnalyzer) Find(name string, pos int) Symbol {
	// Find local symbols first
	path := an.context.MainDocumentPath()
	tree := an.context.Symbols[path]

	if pos != -1 {
		// go innerwards first
		for len(tree.Scopes) != 0 {
			found := false
			for _, s := range tree.Scopes {
				if pos >= s.StartPos.Index && pos <= s.EndPos.Index {
					found = true
					tree = s
					break
				}
			}
			if !found {
				break
			}
		}
	}

	if tree != nil {
		parent := tree

		// search innerwards first then outside
		for parent != nil {
			if sym := parent.Find(name); sym != nil {
				return sym
			} else {
				parent = tree.Parent
			}
		}
	}

	return an.context.FindSymbol(name)
}

type Language struct {
	isCompiled        bool
	Name              string
	FilePatterns      []string
	SitterLanguage    *sitter.Language
	StackTracePattern *regexp.Regexp
	SymbolsToCapture  []SymbolCapture
	LocationConverter func(path, pos string) Location
	ValueAnalyzer     NodeValueAnalyzerFn
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

		contextData.MainError = MainError{
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

	if len(errMsg) == 0 {
		os.Exit(1)
	}

	fmt.Println(Analyze(wd, errMsg))
}
