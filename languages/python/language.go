package python

import (
	"context"
	_ "embed"

	lib "github.com/nedpals/errgoengine"
	"github.com/smacker/go-tree-sitter/python"
)

//go:embed symbols.txt
var symbols string

var Language = &lib.Language{
	Name:              "Python",
	FilePatterns:      []string{".py"},
	SitterLanguage:    python.GetLanguage(),
	StackTracePattern: `\s+File "(?P<path>\S+)", line (?P<position>\d+)(?:, in (?P<symbol>\S+))?`,
	ErrorPattern:      `$stacktrace$message`,
	AnalyzerFactory: func(cd *lib.ContextData) lib.LanguageAnalyzer {
		return &pyAnalyzer{cd}
	},
	SymbolsToCapture: symbols,
}

type pyAnalyzer struct {
	*lib.ContextData
}

func (an *pyAnalyzer) FallbackSymbol() lib.Symbol {
	return BuiltinTypes.AnySymbol
}

func (an *pyAnalyzer) FindSymbol(name string) lib.Symbol {
	sym, _ := builtinTypesStore.FindByName(name)
	return sym
}

func (an *pyAnalyzer) analyzeTypeNode(ctx context.Context, n lib.SyntaxNode) lib.Symbol {
	switch n.Type() {
	case "identifier":
		builtinSym, found := builtinTypesStore.FindByName(n.Text())
		if found {
			return builtinSym
		}
		sym := an.ContextData.FindSymbol(n.Text(), int(n.StartByte()))
		if sym == nil {
			return lib.UnresolvedSymbol
		}
		return sym
	case "subscript":
		valueNode := n.ChildByFieldName("value")
		_, validCollectionType := strToBuiltinCollectionTypeSyms[valueNode.Text()]
		if !validCollectionType {
			baseTypeSym := an.analyzeTypeNode(ctx, valueNode)
			// TODO: inject value syms
			return baseTypeSym
		}

		// get value syms
		valueSyms := make([]lib.Symbol, n.NamedChildCount()-1)
		for i := 1; i < int(n.NamedChildCount()); i++ {
			valueSyms[i-1] = an.AnalyzeNode(ctx, n.NamedChild(i))
		}

		if cSym, err := collectionIfy(valueNode.Text(), valueSyms...); err != nil {
			return cSym
		}
		return BuiltinTypes.AnySymbol
	case "none":
		return BuiltinTypes.NoneSymbol
	}

	// TODO: any or unresolved?
	return BuiltinTypes.AnySymbol
}

func (an *pyAnalyzer) AnalyzeNode(ctx context.Context, n lib.SyntaxNode) lib.Symbol {
	switch n.Type() {
	case "type":
		return an.analyzeTypeNode(ctx, n.NamedChild(0))
	case "true", "false":
		return BuiltinTypes.BooleanSymbol
	case "string":
		return BuiltinTypes.StringSymbol
	case "integer":
		return BuiltinTypes.IntSymbol
	case "float":
		return BuiltinTypes.FloatSymbol
	// case "array_creation_expression":
	// 	var gotLen int
	// 	typeSym := an.AnalyzeNode(n.ChildByFieldName("type"))
	// 	rawLen := n.ChildByFieldName("dimensions").LastNamedChild().Text()
	// 	fmt.Sscanf(rawLen, "%d", &gotLen)
	// 	return arrayIfy(typeSym, gotLen)
	// case "object_creation_expression":
	// 	return an.AnalyzeNode(n.ChildByFieldName("type"))
	case "identifier":
		sym := an.ContextData.FindSymbol(n.Text(), int(n.StartByte()))
		if sym == nil {
			return BuiltinTypes.NoneSymbol
		}
		return sym
	// case "subscript":
	// 	sym := an.AnalyzeNode(n.ChildByFieldName("array"))
	// 	if aSym, ok := sym.(); ok {
	// 		return aSym.ItemSymbol
	// 	} else {
	// 		return BuiltinTypes.VoidSymbol
	// 	}
	case "attribute":
		if objNodeSym := an.AnalyzeNode(ctx, n.ChildByFieldName("object")); objNodeSym != nil {
			if objNodeSym == BuiltinTypes.AnySymbol {
				return objNodeSym
			}

			fieldNode := n.ChildByFieldName("attribute")
			if sym := lib.GetFromSymbol(lib.CastChildrenSymbol(objNodeSym), fieldNode.Text()); sym != nil {
				return sym
			}
		}
	}
	return BuiltinTypes.AnySymbol
}

func (an *pyAnalyzer) AnalyzeImport(params lib.ImportParams) lib.ResolvedImport {
	// TODO:

	return lib.ResolvedImport{
		Path: "",
	}
}
