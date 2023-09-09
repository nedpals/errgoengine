package java

import (
	"fmt"

	lib "github.com/nedpals/errgoengine"
	"github.com/smacker/go-tree-sitter/java"
)

var Language = &lib.Language{
	Name:              "Java",
	FilePatterns:      []string{".java"},
	SitterLanguage:    java.GetLanguage(),
	StackTracePattern: `\s+at (?P<symbol>\S+)\((?P<path>\S+):(?P<position>\d+)\)`,
	ValueAnalyzer: func(an lib.NodeValueAnalyzer, n lib.SyntaxNode) lib.Symbol {
		switch n.Type() {
		// types first
		case "array_type":
			// TODO: types
			return BuiltinTypes.VoidSymbol
		case "boolean_type":
			return BuiltinTypes.BooleanSymbol
		case "void_type":
			return BuiltinTypes.VoidSymbol
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
			typeSym := an.AnalyzeValue(n.ChildByFieldName("type"))
			rawLen := n.ChildByFieldName("dimensions").LastNamedChild().Text()
			fmt.Sscanf(rawLen, "%d", &gotLen)
			return arrayIfy(typeSym, gotLen)
		case "object_creation_expression":
			return an.AnalyzeValue(n.ChildByFieldName("type"))
		case "identifier", "type_identifier":
			if n.Type() == "type_identifier" && n.Text() == "String" {
				return BuiltinTypes.StringSymbol
			}

			sym := an.FindSymbol(n.Text(), int(n.StartByte()))
			if sym == nil {
				return BuiltinTypes.NullSymbol
			}

			return sym
		case "array_access":
			sym := an.AnalyzeValue(n.ChildByFieldName("array"))
			if aSym, ok := sym.(ArraySymbol); ok {
				return aSym.ValueSymbol
			} else {
				return BuiltinTypes.VoidSymbol
			}
		case "field_access", "method_invocation":
			if objNodeSym := an.AnalyzeValue(n.ChildByFieldName("object")); objNodeSym != nil {
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
				return an.AnalyzeValue(parent.ChildByFieldName("type"))
			}
		}
		return BuiltinTypes.VoidSymbol
	},
	ImportResolver: func(an lib.NodeValueAnalyzer, params lib.ImportParams) lib.ResolvedImport {
		// TODO:

		return lib.ResolvedImport{
			Path: "",
		}
	},
	SymbolsToCapture: lib.ISymbolCaptureList{
		lib.SymbolCapture{
			Query: "import_declaration",
			Kind:  lib.SymbolKindImport,
			NameNode: &lib.SymbolCapture{
				Query: `(_) ("." (asterisk))*`,
			},
		},
		lib.SymbolCapture{
			Query: "class_declaration",
			Kind:  lib.SymbolKindClass,
			NameNode: &lib.SymbolCapture{
				Field: "name",
				Query: "identifier",
			},
			BodyNode: &lib.SymbolCapture{
				Field: "body",
				Query: "class_body",
				Children: []*lib.SymbolCapture{
					{
						Query: "field_declaration (variable_declarator)",
						Kind:  lib.SymbolKindVariable,
						NameNode: &lib.SymbolCapture{
							Field: "name",
							Query: "identifier",
						},
					},
					{
						Query: "method_declaration",
						Kind:  lib.SymbolKindFunction,
						NameNode: &lib.SymbolCapture{
							Field: "name",
							Query: "identifier",
						},
						// TODO: make return type node work
						ReturnTypeNode: &lib.SymbolCapture{
							Field: "type",
							Query: "_",
						},
						ContentNode: &lib.SymbolCapture{
							Query:    "return_statement (expression)?",
							Optional: true,
						},
						ParameterNodes: &lib.SymbolCapture{
							Field: "parameters",
							Query: "formal_parameters",
							Children: []*lib.SymbolCapture{
								{
									Kind:  lib.SymbolKindVariable,
									Query: "formal_parameter",
									NameNode: &lib.SymbolCapture{
										Field: "name",
										Query: "identifier",
									},
								},
							},
						},
						BodyNode: &lib.SymbolCapture{
							Field: "body",
							Query: "block",
							Children: []*lib.SymbolCapture{
								// FIXME: figure out return type of variables
								{
									Query: "local_variable_declaration (variable_declarator)",
									Kind:  lib.SymbolKindVariable,
									NameNode: &lib.SymbolCapture{
										Field: "name",
										Query: "identifier",
									},
									ContentNode: &lib.SymbolCapture{
										Field: "value",
										Query: "_",
									},
								},
							},
						},
					},
				},
			},
		},
	},
}
