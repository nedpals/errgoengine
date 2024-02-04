package errgoengine

import (
	"context"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

type symbolTreeKey struct{}

func WithSymbolTree(tree *SymbolTree) context.Context {
	return context.WithValue(context.Background(), symbolTreeKey{}, tree)
}

func GetSymbolTreeCtx(ctx context.Context) *SymbolTree {
	if ctx != nil {
		if tree, ok := ctx.Value(symbolTreeKey{}).(*SymbolTree); ok {
			return tree
		}
	}
	return nil
}

type captureIterator struct {
	idx         int
	checkpoints []int
	doc         *Document
	captures    []sitter.QueryCapture
}

func (it *captureIterator) Get(idx int) sitter.QueryCapture {
	return it.captures[idx]
}

func (it *captureIterator) Reset() {
	it.checkpoints = []int{}
	it.idx = 0
}

func (it *captureIterator) Next() bool {
	if it.idx+1 >= len(it.captures) {
		return false
	}

	it.idx++
	return true
}

func (it *captureIterator) Current() sitter.QueryCapture {
	return it.captures[it.idx]
}

func (it *captureIterator) CurrentNode() SyntaxNode {
	return WrapNode(it.doc, it.Current().Node)
}

func (it *captureIterator) AddIdx(n int) {
	it.idx += n
}

func (it *captureIterator) Rewind() {
	if len(it.checkpoints) == 0 {
		return
	}
	lastIdx := it.checkpoints[0]
	it.checkpoints = it.checkpoints[1:]
	it.idx = lastIdx
}

func (it *captureIterator) GoBack() {
	it.idx--
}

func (it *captureIterator) Save() {
	it.checkpoints = append([]int{it.idx}, it.checkpoints...)
}

type SymbolAnalyzer struct {
	ContextData *ContextData
	doc         *Document
}

func (an *SymbolAnalyzer) analyzeImport(symbolTree *SymbolTree, it *captureIterator) {
	if an.ContextData.Analyzer == nil {
		panic("Node is nil")
	}

	node := it.CurrentNode()
	resolvedImport := an.ContextData.Analyzer.AnalyzeImport(ImportParams{
		Node:       node,
		CurrentDir: an.ContextData.WorkingPath,
	})

	if len(resolvedImport.Path) == 0 {
		// TODO: error
		// return true
		return
	}

	an.ContextData.DepGraph.Add(
		an.ContextData.CurrentDocumentPath,
		map[string]string{
			resolvedImport.Name: resolvedImport.Path,
		})

	symbolTree.Add(&ImportSymbol{
		Alias:           resolvedImport.Name,
		Node:            an.ContextData.DepGraph[resolvedImport.Path],
		ImportedSymbols: resolvedImport.Symbols,
	})
}

func (an *SymbolAnalyzer) analyzeParameters(symbolTree *SymbolTree, query *sitter.Query, it *captureIterator) {
	pNodes := []map[string]SyntaxNode{}
	paramTagMentionCount := 0

	for it.Next() {
		c := it.Current()
		tag := query.CaptureNameForId(c.Index)
		if tag == "parameters" {
			continue
		}

		if tag == "parameter" {
			if paramTagMentionCount < len(pNodes) {
				paramTagMentionCount++
				continue
			}

			// start a new set of parameter node
			paramTagMentionCount++
			pNodes = append(pNodes, map[string]SyntaxNode{})
			continue
		} else if !strings.HasPrefix(tag, "parameter.") {
			it.GoBack()
			break
		}

		tag = strings.TrimPrefix(tag, "parameter.")
		node := it.CurrentNode()

		if paramTagMentionCount == 0 || paramTagMentionCount < len(pNodes) {
			pNodes = append(pNodes, map[string]SyntaxNode{})
		}

		idx := paramTagMentionCount
		if len(pNodes) > 0 && paramTagMentionCount == len(pNodes) {
			idx--
		}

		pNodes[idx][tag] = node
	}

	for _, nodes := range pNodes {
		returnType := an.ContextData.Analyzer.FallbackSymbol()
		if retTypeNode, ok := nodes["return-type"]; ok {
			returnType = an.ContextData.Analyzer.AnalyzeNode(WithSymbolTree(symbolTree), retTypeNode)
		}

		symbolTree.Add(&VariableSymbol{
			Name_:       nodes["name"].Text(),
			Location_:   nodes["name"].Parent().Location(),
			ReturnType_: returnType,
			isParam:     true,
		})
	}
}

func (an *SymbolAnalyzer) analyzeVariables(symbolTree *SymbolTree, query *sitter.Query, it *captureIterator) {
	nodes := map[string]SyntaxNode{}
	nameNodes := []SyntaxNode{}
	var contentReturnType Symbol = an.ContextData.Analyzer.FallbackSymbol()

	for it.Next() {
		c := it.Current()
		tag := query.CaptureNameForId(c.Index)
		if tag == "variable" {
			continue
		} else if !strings.HasPrefix(tag, "variable.") {
			break
		}

		tag = strings.TrimPrefix(tag, "variable.")
		node := it.CurrentNode()
		if tag == "name" {
			nameNodes = append(nameNodes, node)
		} else if tag == "content" {
			contentReturnType = an.ContextData.Analyzer.AnalyzeNode(context.Background(), node)
		} else {
			nodes[tag] = node
		}
	}

	if len(nameNodes) > 0 {
		returnType := an.ContextData.Analyzer.FallbackSymbol()
		if retTypeNode, ok := nodes["return-type"]; ok {
			returnType = an.ContextData.Analyzer.AnalyzeNode(WithSymbolTree(symbolTree), retTypeNode)
		}

		for _, nameNode := range nameNodes {
			symbolTree.Add(&VariableSymbol{
				Name_:             nameNode.Text(),
				Location_:         nameNode.Parent().Location(),
				ReturnType_:       returnType,
				contentReturnType: contentReturnType,
			})
		}
	}
}

func (an *SymbolAnalyzer) analyzeAssignment(symbolTree *SymbolTree, query *sitter.Query, it *captureIterator) {
	nameNodes := []SyntaxNode{}
	contentNodes := []SyntaxNode{}

	for it.Next() {
		c := it.Current()
		tag := query.CaptureNameForId(c.Index)
		if tag == "assignment" {
			continue
		} else if !strings.HasPrefix(tag, "assignment.") {
			break
		}

		tag = strings.TrimPrefix(tag, "assignment.")
		node := it.CurrentNode()
		if tag == "name" {
			nameNodes = append(nameNodes, node)
		} else if tag == "content" {
			contentNodes = append(contentNodes, node)
		}
	}

	if len(nameNodes) > 0 {
		for idx, nameNode := range nameNodes {
			symbolTree.Add(&AssignmentSymbol{
				Variable:     symbolTree.Find(nameNode.Text()),
				FallbackName: nameNode.Text(),
				Location_:    nameNode.Location(),
				ContentReturnType: an.ContextData.Analyzer.AnalyzeNode(
					context.Background(), contentNodes[idx]),
			})
		}
	}
}

func (an *SymbolAnalyzer) analyzeFunction(pre string, symbolTree *SymbolTree, query *sitter.Query, it *captureIterator) {
	parent := it.CurrentNode()
	childTree := symbolTree.CreateChildFromNode(parent)
	nodes := map[string]SyntaxNode{}
	prefix := pre + "."
	returnType := an.ContextData.Analyzer.FallbackSymbol()

	for it.Next() {
		c := it.Current()
		tag := query.CaptureNameForId(c.Index)
		if !strings.HasPrefix(tag, prefix) {
			if tag == "parameters" {
				an.analyzeParameters(childTree, query, it)
				continue
			} else if tag == "block" {
				returnType = an.analyzeBlock(childTree, query, it)
				continue
			} else {
				break
			}
		}
		tag = strings.TrimPrefix(tag, prefix)
		node := it.CurrentNode()

		switch tag {
		case "name":
			nodes[tag] = node
		case "return-type":
			returnType = an.ContextData.Analyzer.AnalyzeNode(WithSymbolTree(childTree), node)
		}
	}

	symbolTree.Add(&TopLevelSymbol{
		Name_:       nodes["name"].Text(),
		Kind_:       SymbolKindFunction,
		Location_:   parent.Location(),
		Children_:   childTree,
		ReturnType_: returnType,
	})
}

func (an *SymbolAnalyzer) analyzeClass(symbolTree *SymbolTree, query *sitter.Query, it *captureIterator) {
	parent := it.CurrentNode()
	nodes := map[string]SyntaxNode{}
	var childTree *SymbolTree

	for it.Next() {
		c := it.Current()
		tag := query.CaptureNameForId(c.Index)

		if !strings.HasPrefix(tag, "class.") {
			an.analyzeUnknown(childTree, query, it)
			break
		}

		tag = strings.TrimPrefix(tag, "class.")
		node := it.CurrentNode()

		switch tag {
		case "name":
			nodes[tag] = node
		case "body":
			childTree = symbolTree.CreateChildFromNode(node)
		}
	}

	symbolTree.Add(&TopLevelSymbol{
		Name_:     nodes["name"].Text(),
		Kind_:     SymbolKindClass,
		Location_: parent.Location(),
		Children_: childTree,
	})
}

func (an *SymbolAnalyzer) analyzeBlock(symbolTree *SymbolTree, query *sitter.Query, it *captureIterator) Symbol {
	nodes := map[string]SyntaxNode{}

	for it.Next() {
		c := it.Current()
		tag := query.CaptureNameForId(c.Index)

		if !strings.HasPrefix(tag, "block.") {
			an.analyzeUnknown(symbolTree, query, it)
			continue
		}

		tag = strings.TrimPrefix(tag, "block.")
		node := it.CurrentNode()
		if tag == "content" {
			nodes[tag] = node
		}
	}

	if contentNode, ok := nodes["content"]; ok {
		return an.ContextData.Analyzer.AnalyzeNode(
			WithSymbolTree(symbolTree), contentNode)
	}

	return an.ContextData.Analyzer.FallbackSymbol()
}

func (an *SymbolAnalyzer) analyzeUnknown(nearest *SymbolTree, query *sitter.Query, it *captureIterator) {
	tag := query.CaptureNameForId(it.Current().Index)
	switch tag {
	case "import":
		an.analyzeImport(nearest, it)
	case "class":
		an.analyzeClass(nearest, query, it)
	case "function", "method":
		an.analyzeFunction(tag, nearest, query, it)
	case "assignment":
		an.analyzeAssignment(nearest, query, it)
	case "block":
		an.analyzeBlock(nearest, query, it)
	case "variable":
		an.analyzeVariables(nearest, query, it)
	}
}

func (an *SymbolAnalyzer) captureAndAnalyze(parent *SymbolTree, rootNode SyntaxNode, symbolCaptures string) {
	if len(symbolCaptures) == 0 {
		return
	} else if parent == nil {
		panic("Parent is null")
	}

	for q := rootNode.Query(symbolCaptures); q.NextMatch(); {
		if q.Len() <= 1 {
			continue
		}

		it := &captureIterator{doc: an.doc, captures: q.Match().Captures()}
		it.Reset()

		nearest := parent.GetNearestScopedTree(int(it.Get(0).Node.StartByte()))
		an.analyzeUnknown(nearest, q.Query(), it)
	}
}

func (an *SymbolAnalyzer) Analyze(doc *Document) {
	oldCurrentDocumentPath := an.ContextData.CurrentDocumentPath

	an.doc = doc
	rootNode := doc.RootNode()
	symTree := an.ContextData.InitOrGetSymbolTree(an.doc.Path)
	an.ContextData.CurrentDocumentPath = an.doc.Path
	an.captureAndAnalyze(symTree, rootNode, an.doc.Language.SymbolsToCapture)
	an.ContextData.CurrentDocumentPath = oldCurrentDocumentPath
}
