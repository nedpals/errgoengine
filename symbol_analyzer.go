package errgoengine

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

type SymbolAnalyzer struct {
	ContextData *ContextData
	doc         *Document
}

func (an *SymbolAnalyzer) analyzeImport(symbolTree *SymbolTree, captures []sitter.QueryCapture) {
	if an.ContextData.Analyzer == nil {
		panic("Node is nil")
	}

	node := WrapNode(an.doc, captures[0].Node)
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

func (an *SymbolAnalyzer) analyzeParameters(symbolTree *SymbolTree, query *sitter.Query, captures []sitter.QueryCapture) {
	nodes := map[string]SyntaxNode{}

	for i, c := range captures[1:] {
		tag := query.CaptureNameForId(c.Index)
		if tag == "parameter" {
			if i == 0 {
				continue
			} else {
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
		} else if !strings.HasPrefix(tag, "parameter.") {
			break
		}

		tag = strings.TrimPrefix(tag, "parameter.")
		node := WrapNode(an.doc, c.Node)
		if tag == "name" || tag == "return-type" {
			nodes[tag] = node
		}
	}
}

func (an *SymbolAnalyzer) analyzeVariables(symbolTree *SymbolTree, query *sitter.Query, captures []sitter.QueryCapture) {
	nodes := map[string]SyntaxNode{}
	nameNodes := []SyntaxNode{}

	for _, c := range captures[1:] {
		tag := query.CaptureNameForId(c.Index)
		if tag == "variable" {
			continue
		} else if !strings.HasPrefix(tag, "variable.") {
			break
		}

		tag = strings.TrimPrefix(tag, "variable.")
		node := WrapNode(an.doc, c.Node)
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

func (an *SymbolAnalyzer) analyzeAssignment(symbolTree *SymbolTree, query *sitter.Query, captures []sitter.QueryCapture) {
	nameNodes := []SyntaxNode{}
	contentNodes := []SyntaxNode{}

	for _, c := range captures[1:] {
		tag := query.CaptureNameForId(c.Index)
		if tag == "assignment" {
			continue
		} else if !strings.HasPrefix(tag, "assignment.") {
			break
		}

		tag = strings.TrimPrefix(tag, "assignment.")
		node := WrapNode(an.doc, c.Node)
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

func (an *SymbolAnalyzer) analyzeFunction(symbolTree *SymbolTree, pre string, query *sitter.Query, captures []sitter.QueryCapture) {
	parent := WrapNode(an.doc, captures[0].Node)
	childTree := symbolTree.CreateChildFromNode(parent)
	nodes := map[string]SyntaxNode{}
	prefix := pre + "."

	for i, c := range captures[1:] {
		tag := query.CaptureNameForId(c.Index)
		if !strings.HasPrefix(tag, prefix) {
			if tag == "parameters" {
				an.analyzeParameters(childTree, query, captures[i+1:])
			} else {
				break
			}
		}

		tag = strings.TrimPrefix(tag, prefix)
		node := WrapNode(an.doc, c.Node)

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

func (an *SymbolAnalyzer) analyzeClass(symbolTree *SymbolTree, query *sitter.Query, captures []sitter.QueryCapture) {
	parent := WrapNode(an.doc, captures[0].Node)
	nodes := map[string]SyntaxNode{}
	var childTree *SymbolTree

	for i, c := range captures[1:] {
		tag := query.CaptureNameForId(c.Index)

		if !strings.HasPrefix(tag, "class.") {
			switch tag {
			case "method":
				an.analyzeFunction(childTree, "method", query, captures[i+1:])
			case "variable":
				an.analyzeVariables(childTree, query, captures[i+1:])
			}

			break
		}

		tag = strings.TrimPrefix(tag, "class.")
		node := WrapNode(an.doc, c.Node)

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

func (an *SymbolAnalyzer) analyzeBlock(symbolTree *SymbolTree, query *sitter.Query, captures []sitter.QueryCapture) {
	nodes := map[string]SyntaxNode{}
	for i, c := range captures[1:] {
		tag := query.CaptureNameForId(c.Index)

		if !strings.HasPrefix(tag, "block.") {
			switch tag {
			case "assignment":
				an.analyzeAssignment(symbolTree, query, captures[i:])
			case "variable":
				an.analyzeVariables(symbolTree, query, captures[i:])
			}

			break
		}

		tag = strings.TrimPrefix(tag, "block.")
		node := WrapNode(an.doc, c.Node)
		if tag == "content" {
			nodes[tag] = node
		}
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

		tag := ctx.Query.CaptureNameForId(ctx.Match.Captures[0].Index)
		switch tag {
		case "import":
			an.analyzeImport(parent, ctx.Match.Captures)
		case "class":
			an.analyzeClass(parent, ctx.Query, ctx.Match.Captures)
		case "function":
			an.analyzeFunction(parent, "function", ctx.Query, ctx.Match.Captures)
		case "assignment":
			nearest := parent.GetNearestScopedTree(int(ctx.Match.Captures[0].Node.StartByte()))
			an.analyzeAssignment(nearest, ctx.Query, ctx.Match.Captures)
		case "variable":
			an.analyzeVariables(parent, ctx.Query, ctx.Match.Captures)
		case "block":
			nearest := parent.GetNearestScopedTree(int(ctx.Match.Captures[0].Node.StartByte()))
			an.analyzeBlock(nearest, ctx.Query, ctx.Match.Captures)
		}
		return true
	})
}

func (an *SymbolAnalyzer) Analyze(doc *Document) {
	an.doc = doc
	rootNode := doc.Tree.RootNode()
	symTree := an.ContextData.InitOrGetSymbolTree(an.doc.Path)
	an.ContextData.CurrentDocumentPath = an.doc.Path
	an.captureAndAnalyze(symTree, WrapNode(doc, rootNode), an.doc.Language.SymbolsToCapture)
	an.ContextData.CurrentDocumentPath = ""
}
