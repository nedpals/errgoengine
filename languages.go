package main

import "github.com/smacker/go-tree-sitter/java"

var JavaLanguage = &Language{
	Name:           "Java",
	FilePatterns:   []string{".java"},
	SitterLanguage: java.GetLanguage(),
	BuiltinTypes:   []string{"null", "void", "string", "char", "int", "float", "double"},
	SymbolExtractors: map[string]SymbolExtractorFn{
		"class_declaration": func(n Node, an *Analyzer) {
			nameNode := n.ChildByFieldName("name")
			bodyNode := n.ChildByFieldName("body")

			sym := an.AddSymbolFromWrappedNode(nameNode, SymbolTypeClass, n.Location())
			an.SetParent(sym)
			an.AnalyzeNode(bodyNode.Node)
		},
		"method_declaration": func(n Node, an *Analyzer) {
			nameNode := n.ChildByFieldName("name")
			sym := an.AddSymbolFromWrappedNode(nameNode, SymbolTypeFunction, n.Location())
			an.SetParent(sym)

			bodyNode := n.ChildByFieldName("body")
			an.AnalyzeNode(bodyNode.Node)
		},
		"local_variable_declaration": func(n Node, an *Analyzer) {
			declaratorNode := n.ChildByFieldName("declarator")
			an.AnalyzeNode(declaratorNode.Node)
		},
		"variable_declarator": func(n Node, an *Analyzer) {
			// typeNode := n.ChildByFieldName("type")
			nameNode := n.ChildByFieldName("name")
			an.AddSymbol(&Symbol{
				Name: nameNode.Text(),
				Type: SymbolTypeVariable,
				// TODO ReturnSymbol: ,
			})
		},
	},
}
