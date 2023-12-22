package java

import (
	_ "embed"
	"fmt"

	lib "github.com/nedpals/errgoengine"
	"github.com/smacker/go-tree-sitter/java"
)

//go:embed symbols.txt
var symbols string

var Language = &lib.Language{
	Name:              "Java",
	FilePatterns:      []string{".java"},
	SitterLanguage:    java.GetLanguage(),
	StackTracePattern: `\s+at (?P<symbol>\S+)\((?P<path>\S+):(?P<position>\d+)\)`,
	AnalyzerFactory: func(cd *lib.ContextData) lib.LanguageAnalyzer {
		return &javaAnalyzer{cd}
	},
	SymbolsToCapture: symbols,
}

type javaAnalyzer struct {
	*lib.ContextData
}

func (an *javaAnalyzer) FallbackSymbol() lib.Symbol {
	return BuiltinTypes.VoidSymbol
}

func (an *javaAnalyzer) AnalyzeNode(n lib.SyntaxNode) lib.Symbol {
	switch n.Type() {
	// types first
	case "array_type":
		// TODO: types
		return BuiltinTypes.VoidSymbol
	case "boolean_type":
		return BuiltinTypes.BooleanSymbol
	case "void_type":
		return BuiltinTypes.VoidSymbol
	case "integral_type":
		return BuiltinTypes.Integral.IntSymbol
	// then expressions
	case "null_literal":
		return BuiltinTypes.NullSymbol
	case "true", "false":
		return BuiltinTypes.BooleanSymbol
	case "string_literal":
		return BuiltinTypes.StringSymbol
	case "character_literal":
		return BuiltinTypes.Integral.CharSymbol
	case "octal_integer_literal",
		"hex_integer_literal",
		"binary_integer_literal":
		return BuiltinTypes.Integral.IntSymbol
	case "decimal_floating_point_literal",
		"hex_floating_point_literal":
		// TODO: value check if float or double
		return BuiltinTypes.FloatingPoint.DoubleSymbol
	case "array_creation_expression":
		var gotLen int
		typeSym := an.AnalyzeNode(n.ChildByFieldName("type"))
		rawLen := n.ChildByFieldName("dimensions").LastNamedChild().Text()
		fmt.Sscanf(rawLen, "%d", &gotLen)
		return arrayIfy(typeSym, gotLen)
	case "object_creation_expression":
		return an.AnalyzeNode(n.ChildByFieldName("type"))
	case "identifier", "type_identifier":
		if n.Type() == "type_identifier" && n.Text() == "String" {
			return BuiltinTypes.StringSymbol
		}

		sym := an.FindSymbol(n.Text(), int(n.StartByte()))
		if sym == nil {
			if n.Type() == "type_identifier" {
				an.ContextData.

				// mark type as unresolved
				return lib.UnresolvedSymbol
			}
			return BuiltinTypes.NullSymbol
		}

		return sym
	case "array_access":
		sym := an.AnalyzeNode(n.ChildByFieldName("array"))
		if aSym, ok := sym.(ArraySymbol); ok {
			return aSym.ValueSymbol
		} else {
			return BuiltinTypes.VoidSymbol
		}
	case "field_access", "method_invocation":
		if objNodeSym := an.AnalyzeNode(n.ChildByFieldName("object")); objNodeSym != nil {
			if objNodeSym == BuiltinTypes.NullSymbol {
				return objNodeSym
			}

			if n.Type() == "field_access" {
				fieldNode := n.ChildByFieldName("field")
				if sym := lib.GetFromSymbol(lib.CastChildrenSymbol(objNodeSym), fieldNode.Text()); sym != nil {
					return sym
				}
			}
		}
	case "this":
		// TODO: support this
		return BuiltinTypes.VoidSymbol
	case "block":
		if parent := n.Parent(); parent.Type() == "method_declaration" {
			return an.AnalyzeNode(parent.ChildByFieldName("type"))
		}
	}
	return BuiltinTypes.VoidSymbol
}

func (an *javaAnalyzer) AnalyzeImport(params lib.ImportParams) lib.ResolvedImport {
	// TODO:

	return lib.ResolvedImport{
		Path: "",
	}
}
