package java

import (
	"context"
	"embed"
	_ "embed"
	"fmt"

	lib "github.com/nedpals/errgoengine"
	"github.com/smacker/go-tree-sitter/java"
)

//go:embed symbols.txt
var symbols string

//go:embed externs/**/*.json
var externs embed.FS

var Language = &lib.Language{
	Name:              "Java",
	FilePatterns:      []string{".java"},
	SitterLanguage:    java.GetLanguage(),
	StackTracePattern: `\s+at (?P<symbol>\S+)\((?P<path>\S+):(?P<position>\d+)\)`,
	AnalyzerFactory: func(cd *lib.ContextData) lib.LanguageAnalyzer {
		return &javaAnalyzer{cd}
	},
	SymbolsToCapture: symbols,
	ExternFS:         externs,
	LocationConverter: func(ctx lib.LocationConverterContext) lib.Location {
		var trueLine int
		if _, err := fmt.Sscanf(ctx.Pos, "%d", &trueLine); err != nil {
			panic(err)
		}
		return lib.Location{
			DocumentPath: ctx.Path,
			StartPos:     lib.Position{Line: trueLine},
			EndPos:       lib.Position{Line: trueLine},
		}
	},
}

type javaAnalyzer struct {
	*lib.ContextData
}

func (an *javaAnalyzer) FallbackSymbol() lib.Symbol {
	return BuiltinTypes.VoidSymbol
}

func (an *javaAnalyzer) FindSymbol(name string) lib.Symbol {
	sym, _ := builtinTypesStore.FindByName(name)
	return sym
}

func (an *javaAnalyzer) AnalyzeNode(ctx context.Context, n lib.SyntaxNode) lib.Symbol {
	symbolTree := lib.GetSymbolTreeCtx(ctx)

	switch n.Type() {
	// types first
	case "array_type":
		// TODO: types
		dimNode := n.ChildByFieldName("dimensions")
		len := 0
		if dimNode.NamedChildCount() != 0 {
			len, _ = fmt.Sscanf(dimNode.FirstNamedChild().Text(), "%d", &len)
		}

		elNode := n.ChildByFieldName("element")
		elSym := an.AnalyzeNode(ctx, elNode)
		return arrayIfy(elSym, len)
	case "void_type":
		return BuiltinTypes.VoidSymbol
	case "type_identifier", "boolean_type", "integral_type", "floating_point_type":
		// check for builtin types first
		builtinSym, found := builtinTypesStore.FindByName(n.Text())
		if found {
			return builtinSym
		}
		sym := an.ContextData.FindSymbol(n.Text(), int(n.StartByte()))
		if sym == nil {
			return lib.UnresolvedSymbol
		}
		return sym
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
		typeSym := an.AnalyzeNode(ctx, n.ChildByFieldName("type"))
		rawLen := n.ChildByFieldName("dimensions").LastNamedChild().Text()
		fmt.Sscanf(rawLen, "%d", &gotLen)
		return arrayIfy(typeSym, gotLen)
	case "object_creation_expression":
		return an.AnalyzeNode(ctx, n.ChildByFieldName("type"))
	case "identifier":
		sym := an.ContextData.FindSymbol(n.Text(), int(n.StartByte()))
		if sym == nil && symbolTree != nil {
			sym = symbolTree.Find(n.Text())
		}
		if sym == nil {
			return BuiltinTypes.NullSymbol
		}

		return sym
	case "array_access":
		sym := an.AnalyzeNode(ctx, n.ChildByFieldName("array"))
		if aSym, ok := sym.(ArraySymbol); ok {
			return aSym.ItemSymbol
		} else {
			return BuiltinTypes.VoidSymbol
		}
	case "field_access", "method_invocation":
		objNodeSym := an.FallbackSymbol()

		if objNode := n.ChildByFieldName("object"); !objNode.IsNull() {
			objNodeSym = an.AnalyzeNode(ctx, n.ChildByFieldName("object"))
			if objNodeSym == BuiltinTypes.NullSymbol {
				return objNodeSym
			}
		}

		if n.Type() == "field_access" {
			fieldNode := n.ChildByFieldName("field")
			var sym lib.Symbol

			if objNodeSym == an.FallbackSymbol() {
				sym = an.ContextData.FindSymbol(fieldNode.Text(), int(fieldNode.StartByte()))
			} else {
				sym = lib.GetFromSymbol(lib.CastChildrenSymbol(objNodeSym), fieldNode.Text())
			}

			if sym == nil {
				return sym
			}
		} else if n.Type() == "method_invocation" {
			nameNode := n.ChildByFieldName("name")
			var sym lib.Symbol

			if objNodeSym == an.FallbackSymbol() {
				sym = an.ContextData.FindSymbol(nameNode.Text(), int(nameNode.StartByte()))
			} else {
				sym = lib.GetFromSymbol(lib.CastChildrenSymbol(objNodeSym), nameNode.Text())
			}

			if sym == nil {
				return BuiltinTypes.VoidSymbol
			}

			methodSym, ok := sym.(*lib.TopLevelSymbol)
			if !ok || sym.Kind() != lib.SymbolKindFunction {
				return BuiltinTypes.VoidSymbol
			}

			// input to parameter type check
			argumentsNode := n.ChildByFieldName("arguments")
			argsIdx := 0

			for _, paramSym := range methodSym.Children().Symbols {
				if argsIdx >= int(argumentsNode.NamedChildCount()) {
					break
				}

				paramSym, ok := paramSym.(*lib.VariableSymbol)
				if !ok || !paramSym.IsParam() {
					continue
				}

				argNode := argumentsNode.NamedChild(argsIdx)
				argSym := lib.UnwrapReturnType(an.AnalyzeNode(ctx, argNode))
				if argSym != paramSym.ReturnType() {
					return BuiltinTypes.VoidSymbol
				}

				argsIdx++
			}

			return methodSym.ReturnType()
		}
	case "this":
		// TODO: support this
		return BuiltinTypes.VoidSymbol
	case "block":
		if parent := n.Parent(); parent.Type() == "method_declaration" {
			return an.AnalyzeNode(ctx, parent.ChildByFieldName("type"))
		}
	case "binary_expression":
		leftSym := lib.UnwrapReturnType(an.AnalyzeNode(ctx, n.ChildByFieldName("left")))
		rightSym := lib.UnwrapReturnType(an.AnalyzeNode(ctx, n.ChildByFieldName("right")))
		if leftSym == rightSym {
			return leftSym
		}
		return BuiltinTypes.VoidSymbol
	}
	return BuiltinTypes.VoidSymbol
}

func (an *javaAnalyzer) AnalyzeImport(params lib.ImportParams) lib.ResolvedImport {
	// TODO:

	return lib.ResolvedImport{
		Path: "",
	}
}
