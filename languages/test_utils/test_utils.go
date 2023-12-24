package testutils

import (
	"fmt"
	"strings"
	"testing"

	lib "github.com/nedpals/errgoengine"
	testutils "github.com/nedpals/errgoengine/test_utils"
	sitter "github.com/smacker/go-tree-sitter"
)

type TestCase struct {
	Name     string
	FileName string
	Input    string
	Expected string
}

type TestCases []TestCase

func (cases TestCases) Execute(t *testing.T, lang *lib.Language) {
	lang.Compile()

	parser := sitter.NewParser()
	analyzer := &lib.SymbolAnalyzer{
		ContextData: lib.NewContextData(lib.NewEmptyStore(), ""),
	}

	analyzer.ContextData.Analyzer = lang.AnalyzerFactory(analyzer.ContextData)
	sb := &strings.Builder{}

	for _, tCase := range cases {
		t.Run(tCase.Name, func(t *testing.T) {
			doc, err := lib.ParseDocument(tCase.FileName, strings.NewReader(strings.TrimSpace(tCase.Input)), parser, lang, nil)
			if err != nil {
				t.Fatal(err)
			}

			analyzer.Analyze(doc)
			treeSexprBuilder(analyzer.ContextData.Symbols[doc.Path], sb, 0)
			defer sb.Reset()

			testutils.Equals(t, sb.String(), strings.TrimSpace(tCase.Expected))
		})
	}
}

func treeSexprBuilder(tree *lib.SymbolTree, sb *strings.Builder, level int) {
	sb.WriteString(strings.Repeat("\t", level))
	sb.WriteByte('(')
	sb.WriteString(fmt.Sprintf("tree %s-%s", tree.StartPos, tree.EndPos))
	if len(tree.Symbols) > 0 {
		sb.WriteByte('\n')
		i := 0
		for _, sym := range tree.Symbols {
			if i != 0 && i < len(tree.Symbols) {
				sb.WriteByte('\n')
			}
			sb.WriteString(strings.Repeat("\t", level+1))
			sb.WriteByte('(')
			sb.WriteString(sym.Kind().String())
			if sym, ok := sym.(lib.IReturnableSymbol); ok {
				sb.WriteString(fmt.Sprintf(" %s", sym.ReturnType().Name()))
			}
			sb.WriteString(fmt.Sprintf(" %s %s-%s", sym.Name(), sym.Location().StartPos, sym.Location().EndPos))

			if childSym, ok := sym.(lib.IChildrenSymbol); ok && childSym.Children() != nil {
				sb.WriteByte('\n')
				treeSexprBuilder(childSym.Children(), sb, level+2)
			}

			sb.WriteByte(')')
			i++
		}
	}
	sb.WriteByte(')')
}
