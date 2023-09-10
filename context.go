package errgoengine

import sitter "github.com/smacker/go-tree-sitter"

type MainError struct {
	ErrorNode *StackTraceEntry
	Document  *Document
	Nearest   SyntaxNode
}

func (err MainError) DocumentPath() string {
	return err.ErrorNode.DocumentPath
}

type ContextData struct {
	*Store
	Analyzer            LanguageAnalyzer
	WorkingPath         string
	CurrentDocumentPath string
	Variables           map[string]string
	TraceStack          TraceStack
	MainError           MainError
}

func NewContextData(store *Store, workingPath string) *ContextData {
	return &ContextData{
		Store:       store,
		WorkingPath: workingPath,
		Variables:   make(map[string]string),
		TraceStack:  TraceStack{},
	}
}

func (data *ContextData) MainDocumentPath() string {
	if data.MainError.ErrorNode != nil {
		return data.MainError.DocumentPath()
	}
	return data.CurrentDocumentPath
}

func (data *ContextData) FindSymbol(name string, pos int) Symbol {
	path := data.MainDocumentPath()
	return data.Store.FindSymbol(path, name, pos)
}

func (data *ContextData) AddVariable(name string, value string) {
	if data.Variables == nil {
		data.Variables = make(map[string]string)
	}

	data.Variables[name] = value
}

func getNearestNodeWithSymbol(data *ContextData, rootNode SyntaxNode, sym Symbol) SyntaxNode {
	cursor := sitter.NewTreeCursor(rootNode.Node)
	cursor.GoToFirstChild()

	defer cursor.Close()

	for {
		if n := cursor.CurrentNode(); n.IsNamed() {
			wrappedNode := WrapNode(rootNode.Doc, n)
			gotSym := data.Analyzer.AnalyzeNode(wrappedNode)
			if gotSym == sym {
				return wrappedNode
			} else if n.NamedChildCount() != 0 {
				gotNode := getNearestNodeWithSymbol(data, wrappedNode, sym)
				if gotNode == nil {
					// TODO:
				}
			}
		}

		if !cursor.GoToNextSibling() {
			break
		}
	}
}

func NearestNodeWithSymbol(data *ContextData, rootNode SyntaxNode, sym Symbol, callback func(*sitter.QueryMatch, *sitter.Query) bool) {

	queryStr := ""
}
