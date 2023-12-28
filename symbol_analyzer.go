package errgoengine

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

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

	for it.Next() {
		c := it.Current()
		tag := query.CaptureNameForId(c.Index)
		if tag == "parameters" {
			continue
		}

		if tag == "parameter" {
			// start a new set of parameter node
			pNodes = append(pNodes, map[string]SyntaxNode{})
			continue
		} else if !strings.HasPrefix(tag, "parameter.") {
			break
		}

		tag = strings.TrimPrefix(tag, "parameter.")
		node := it.CurrentNode()
		pNodes[len(pNodes)-1][tag] = node
	}

	for _, nodes := range pNodes {
		returnType := an.ContextData.Analyzer.FallbackSymbol()
		if retTypeNode, ok := nodes["return-type"]; ok {
			returnType = an.ContextData.Analyzer.AnalyzeNode(retTypeNode)
		}

		symbolTree.Add(&VariableSymbol{
			Name_:       nodes["name"].Text(),
			Location_:   nodes["name"].Parent().Location(),
			ReturnType_: returnType,
		})
	}
}

func (an *SymbolAnalyzer) analyzeVariables(symbolTree *SymbolTree, query *sitter.Query, it *captureIterator) {
	nodes := map[string]SyntaxNode{}
	nameNodes := []SyntaxNode{}

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
		} else {
			nodes[tag] = node
		}
	}

	if len(nameNodes) > 0 {
		returnType := an.ContextData.Analyzer.FallbackSymbol()
		if retTypeNode, ok := nodes["return-type"]; ok {
			returnType = an.ContextData.Analyzer.AnalyzeNode(retTypeNode)
		}

		for _, nameNode := range nameNodes {
			symbolTree.Add(&VariableSymbol{
				Name_:       nameNode.Text(),
				Location_:   nameNode.Parent().Location(),
				ReturnType_: returnType,
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
				Variable:          symbolTree.Find(nameNode.Text()),
				FallbackName:      nameNode.Text(),
				Location_:         nameNode.Location(),
				ContentReturnType: an.ContextData.Analyzer.AnalyzeNode(contentNodes[idx]),
			})
		}
	}
}

func (an *SymbolAnalyzer) analyzeFunction(pre string, symbolTree *SymbolTree, query *sitter.Query, it *captureIterator) {
	parent := it.CurrentNode()
	childTree := symbolTree.CreateChildFromNode(parent)
	nodes := map[string]SyntaxNode{}
	prefix := pre + "."

	for it.Next() {
		c := it.Current()
		tag := query.CaptureNameForId(c.Index)
		if !strings.HasPrefix(tag, prefix) {
			if tag == "parameters" {
				an.analyzeParameters(childTree, query, it)
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
		}
	}

	symbolTree.Add(&TopLevelSymbol{
		Name_:     nodes["name"].Text(),
		Kind_:     SymbolKindFunction,
		Location_: parent.Location(),
		Children_: childTree,
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

func (an *SymbolAnalyzer) analyzeBlock(symbolTree *SymbolTree, query *sitter.Query, it *captureIterator) {
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

	QueryNode(rootNode, strings.NewReader(symbolCaptures), func(ctx QueryNodeCtx) bool {
		if len(ctx.Match.Captures) <= 1 || ctx.Match.Captures == nil {
			return true
		}

		it := &captureIterator{doc: an.doc, captures: ctx.Match.Captures}
		it.Reset()

		nearest := parent.GetNearestScopedTree(int(it.Get(0).Node.StartByte()))
		an.analyzeUnknown(nearest, ctx.Query, it)

		return true
	})
}

func (an *SymbolAnalyzer) Analyze(doc *Document) {
	an.doc = doc
	rootNode := doc.RootNode()
	symTree := an.ContextData.InitOrGetSymbolTree(an.doc.Path)
	an.ContextData.CurrentDocumentPath = an.doc.Path
	an.captureAndAnalyze(symTree, rootNode, an.doc.Language.SymbolsToCapture)
	an.ContextData.CurrentDocumentPath = ""
}
