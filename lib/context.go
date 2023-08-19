package lib

import (
	sitter "github.com/smacker/go-tree-sitter"
)

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

func (data *ContextData) FindSymbol(name string, pos int) Symbol {
	// Find local symbols first
	path := data.MainDocumentPath()
	tree := data.Symbols[path]

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

	return nil
}

func (data *ContextData) AnalyzeValue(n Node) Symbol {
	return n.Doc.Language.ValueAnalyzer(data, n)
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
