package main

import (
	"regexp"
	"strconv"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/java"
)

var nullSymbol = BuiltinSymbol("null")
var booleanSymbol = BuiltinSymbol("boolean")
var stringSymbol = BuiltinSymbol("string")
var charSymbol = BuiltinSymbol("char")
var intSymbol = BuiltinSymbol("int")

// var floatSymbol = BuiltinSymbol("float")
var doubleSymbol = BuiltinSymbol("double")
var voidSymbol = BuiltinSymbol("void")

var JavaLanguage = &Language{
	Name:              "Java",
	FilePatterns:      []string{".java"},
	SitterLanguage:    java.GetLanguage(),
	BuiltinTypes:      []string{"null", "boolean", "void", "string", "char", "int", "float", "double"},
	StackTracePattern: regexp.MustCompile(`\s+at (?P<symbol>\S+)\((?P<path>\S+):(?P<position>\d+)\)`),
	LocationConverter: func(path, pos string) Location {
		trueLine, err := strconv.Atoi(pos)
		if err != nil {
			panic(err)
		}
		return Location{
			DocumentPath: path,
			Position:     Position{Line: trueLine},
		}
	},
	ValueAnalyzer: func(an *NodeValueAnalyzer, n Node) *Symbol {
		switch n.Type() {
		case "null_literal":
			return nullSymbol
		case "true", "false":
			return booleanSymbol
		case "string_literal":
			return stringSymbol
		case "character_literal":
			return charSymbol
		case "octal_integer_literal",
			"hex_integer_literal",
			"binary_integer_literal":
			return intSymbol
		case "decimal_floating_point_literal",
			"hex_floating_point_literal":
			// TODO: value check if float or double
			return doubleSymbol
		case "object_creation_expression":
			typeNode := n.ChildByFieldName("type")
			if sym := an.Find(typeNode.Text()); sym != nil {
				return sym
			}
		case "identifier":
			return an.Find(n.Text())
		case "field_access":
			if objNodeSym := an.Analyze(n.ChildByFieldName("object")); objNodeSym != nil {
				if objNodeSym == nullSymbol {
					return voidSymbol
				}

				fieldNode := n.ChildByFieldName("field")
				if sym := objNodeSym.Get(fieldNode.Text()); sym != nil {
					return sym
				}
			}
		case "this":
			// TODO: support this
			return voidSymbol
		}
		return voidSymbol
	},
	ValueNodeTransformer: func(transform ValueNodeTransformer, node *sitter.Node) *sitter.Node {
		// if node.Type() == "expression_statement" {
		// 	return node.Child(0)
		// }
		return node
	},
	SymbolsToCapture: []SymbolCapture{
		{
			Query: "class_declaration",
			Kind:  SymbolKindClass,
			NameNode: &SymbolCapture{
				Field: "name",
				Query: "identifier",
			},
			BodyNode: &SymbolCapture{
				Field: "body",
				Query: "class_body",
				Children: []*SymbolCapture{
					{
						Query: "field_declaration (variable_declarator)",
						Kind:  SymbolKindVariable,
						NameNode: &SymbolCapture{
							Field: "name",
							Query: "identifier",
						},
					},
					{
						Query: "method_declaration",
						Kind:  SymbolKindFunction,
						NameNode: &SymbolCapture{
							Field: "name",
							Query: "identifier",
						},
						// TODO: make return type node work
						ReturnTypeNode: &SymbolCapture{
							Field: "type",
							Query: "_",
						},
						ContentNode: &SymbolCapture{
							Query:    "return_statement (expression)?",
							Optional: true,
						},
						ParameterNodes: &SymbolCapture{
							Field: "parameters",
							Query: "formal_parameters",
							Children: []*SymbolCapture{
								{
									Kind:  SymbolKindVariable,
									Query: "formal_parameter",
									NameNode: &SymbolCapture{
										Field: "name",
										Query: "identifier",
									},
								},
							},
						},
						BodyNode: &SymbolCapture{
							Field: "body",
							Query: "block",
							Children: []*SymbolCapture{
								// FIXME: figure out return type of variables
								{
									Query: "local_variable_declaration (variable_declarator)",
									Kind:  SymbolKindVariable,
									NameNode: &SymbolCapture{
										Field: "name",
										Query: "identifier",
									},
									ContentNode: &SymbolCapture{
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
