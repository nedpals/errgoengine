package main

import (
	"fmt"
	"regexp"

	"github.com/smacker/go-tree-sitter/java"
)

type ArraySymbol struct {
	ValueSymbol Symbol
	Length      int
}

func (sym ArraySymbol) Name() string {
	return fmt.Sprintf("[%d]%s", sym.Length, sym.ValueSymbol.Name())
}

func (sym ArraySymbol) Kind() SymbolKind {
	return SymbolKindArray
}

func (sym ArraySymbol) Location() Location {
	return sym.ValueSymbol.Location()
}

var BuiltinTypes = struct {
	NullSymbol    Symbol
	BooleanSymbol Symbol
	StringSymbol  Symbol
	Integral      struct {
		ByteSymbol  Symbol
		ShortSymbol Symbol
		IntSymbol   Symbol
		LongSymbol  Symbol
		CharSymbol  Symbol
	}
	FloatingPoint struct {
		FloatSymbol  Symbol
		DoubleSymbol Symbol
	}
	VoidSymbol Symbol
}{
	NullSymbol:    BuiltinSymbol{"null"},
	BooleanSymbol: BuiltinSymbol{"boolean"},
	StringSymbol:  BuiltinSymbol{"string"},
	Integral: struct {
		ByteSymbol  Symbol
		ShortSymbol Symbol
		IntSymbol   Symbol
		LongSymbol  Symbol
		CharSymbol  Symbol
	}{
		ByteSymbol:  BuiltinSymbol{"byte"},
		ShortSymbol: BuiltinSymbol{"short"},
		IntSymbol:   BuiltinSymbol{"int"},
		LongSymbol:  BuiltinSymbol{"long"},
		CharSymbol:  BuiltinSymbol{"char"},
	},
	FloatingPoint: struct {
		FloatSymbol  Symbol
		DoubleSymbol Symbol
	}{
		FloatSymbol:  BuiltinSymbol{"float"},
		DoubleSymbol: BuiltinSymbol{"double"},
	},
	VoidSymbol: BuiltinSymbol{"void"},
}

func arrayIfy(typ Symbol, len int) Symbol {
	return ArraySymbol{
		ValueSymbol: typ,
		Length:      len,
	}
}

var JavaLanguage = &Language{
	Name:              "Java",
	FilePatterns:      []string{".java"},
	SitterLanguage:    java.GetLanguage(),
	StackTracePattern: regexp.MustCompile(`\s+at (?P<symbol>\S+)\((?P<path>\S+):(?P<position>\d+)\)`),
	LocationConverter: func(path, pos string) Location {
		var trueLine int
		if _, err := fmt.Sscanf(pos, "%d", &trueLine); err != nil {
			panic(err)
		}
		return Location{
			DocumentPath: path,
			Position:     Position{Line: trueLine},
		}
	},
	ValueAnalyzer: func(an *NodeValueAnalyzer, n Node) Symbol {
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
			typeSym := an.Analyze(n.ChildByFieldName("type"))
			rawLen := n.ChildByFieldName("dimensions").LastNamedChild().Text()
			fmt.Sscanf(rawLen, "%d", &gotLen)
			return arrayIfy(typeSym, gotLen)
		case "object_creation_expression":
			return an.Analyze(n.ChildByFieldName("type"))
		case "identifier", "type_identifier":
			if n.Type() == "type_identifier" && n.Text() == "String" {
				return BuiltinTypes.StringSymbol
			}

			sym := an.Find(n.Text(), int(n.StartByte()))
			if sym == nil {
				return BuiltinTypes.NullSymbol
			}

			return sym
		case "array_access":
			sym := an.Analyze(n.ChildByFieldName("array"))
			if aSym, ok := sym.(ArraySymbol); ok {
				return aSym.ValueSymbol
			} else {
				return BuiltinTypes.VoidSymbol
			}
		case "field_access", "method_invocation":
			if objNodeSym := an.Analyze(n.ChildByFieldName("object")); objNodeSym != nil {
				if objNodeSym == BuiltinTypes.NullSymbol {
					return objNodeSym
				}

				if n.Type() == "field_access" {
					fieldNode := n.ChildByFieldName("field")
					if sym := GetFromSymbol(CastChildrenSymbol(objNodeSym), fieldNode.Text()); sym != nil {
						return sym
					}
				}
			}
		case "this":
			// TODO: support this
			return BuiltinTypes.VoidSymbol
		case "block":
			if parent := n.Parent(); parent.Type() == "method_declaration" {
				return an.Analyze(parent.ChildByFieldName("type"))
			}
		}
		return BuiltinTypes.VoidSymbol
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
