package lib

import (
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

type SymbolAnalyzer struct {
	contextData *ContextData
	doc         *Document
}

func (an *SymbolAnalyzer) captureAndAnalyze(parent *SymbolTree, rootNode *sitter.Node, symbolCaptures ...ISymbolCapture) {
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
			captured[key] = WrapNode(an.doc, c.Node)
			if len(firstMatchCname) == 0 && SymPrefixRegex.MatchString(key) {
				firstMatchCname = key
			}
		}

		if len(captured) == 0 {
			continue
		}

		var identifiedKind SymbolKind
		var captureIdx int

		_, err := fmt.Sscanf(firstMatchCname, SymPrefix, &captureIdx, &identifiedKind)
		if err != nil {
			panic(err)
		}

		// rename map entries
		for k := range captured {
			renamed := strings.TrimPrefix(k, fmt.Sprintf(SymPrefix+".", captureIdx, identifiedKind))
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
				children = SymCaptureToListPtr(symCapture.(SymbolCapture).BodyNode.Children)
			case *SymbolCapture:
				children = SymCaptureToListPtr(symCapture.(*SymbolCapture).BodyNode.Children)
			}

			an.captureAndAnalyze(childTree, body.RawNode(), children...)
			parent.Add(&TopLevelSymbol{
				Name_:     captured["name"].Text(),
				Kind_:     identifiedKind,
				Location_: captured["sym"].Location(),
				Children_: childTree,
			})
		} else if content, ok := captured["content"]; ok {
			returnSym := an.contextData.AnalyzeValue(content)
			parent.Add(&VariableSymbol{
				Name_:       captured["name"].Text(),
				Location_:   captured["sym"].Location(),
				ReturnType_: returnSym,
			})
		}
	}
}

func (an *SymbolAnalyzer) AnalyzeTree(tree *sitter.Tree) {
	rootNode := tree.RootNode()
	symTree := an.contextData.InitOrGetSymbolTree(an.doc.Path)
	an.contextData.CurrentDocumentPath = an.doc.Path
	an.captureAndAnalyze(symTree, rootNode, an.doc.Language.SymbolsToCapture...)
	an.contextData.CurrentDocumentPath = ""
}
