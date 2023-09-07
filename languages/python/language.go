package python

import (
	lib "github.com/nedpals/errgoengine"
	"github.com/smacker/go-tree-sitter/python"
)

var Language = &lib.Language{
	Name:              "Python",
	FilePatterns:      []string{".py"},
	SitterLanguage:    python.GetLanguage(),
	StackTracePattern: `\s+File "(?P<path>\S+)", line (?P<position>\d+), in (?P<symbol>\S+)`,
	ErrorPattern:      `Traceback \(most recent call last\):$stacktrace$message`,
	ValueAnalyzer: func(nva lib.NodeValueAnalyzer, n lib.Node) lib.Symbol {
		// TODO:
		return lib.Builtin("void")
	},
	ImportResolver: func(an lib.NodeValueAnalyzer, params lib.ImportParams) lib.ResolvedImport {
		// TODO:

		return lib.ResolvedImport{
			Path: "",
		}
	},
	SymbolsToCapture: lib.ISymbolCaptureList{
		lib.SymbolCapture{
			Query: "import_statement",
			Kind:  lib.SymbolKindImport,
			NameNode: &lib.SymbolCapture{
				Query: "_",
				Field: "name",
			},
		},
		lib.SymbolCapture{
			Query: "import_from_statement",
			Kind:  lib.SymbolKindImport,
			NameNode: &lib.SymbolCapture{
				Query: "_",
				Field: "module_name",
			},
			// TODO: include symbol names?
		},
		lib.SymbolCapture{
			Query: "function_definition",
			Kind:  lib.SymbolKindFunction,
			NameNode: &lib.SymbolCapture{
				Query: "identifier",
				Field: "name",
			},
			ParameterNodes: &lib.SymbolCapture{
				Field: "parameters",
				Query: "parameters",
				Children: []*lib.SymbolCapture{
					{
						Kind:  lib.SymbolKindVariable,
						Query: "parameter",
						NameNode: &lib.SymbolCapture{
							Query: "identifier",
						},
					},
					{
						Kind:  lib.SymbolKindVariable,
						Query: "parameter",
						NameNode: &lib.SymbolCapture{
							Query: "_",
							Field: "name",
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
						Query: "expression_statement (assignment)",
						Kind:  lib.SymbolKindVariable,
						NameNode: &lib.SymbolCapture{
							Field: "left",
							Query: "identifier",
						},
						ContentNode: &lib.SymbolCapture{
							Field: "right",
							Query: "_",
						},
					},
				},
			},
		},
	},
}
