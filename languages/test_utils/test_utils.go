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
	sb.WriteString(fmt.Sprintf("tree [%d,%d]-[%d,%d]", tree.StartPos.Line, tree.StartPos.Column, tree.EndPos.Line, tree.EndPos.Column))
	if len(tree.Symbols) > 0 {
		sb.WriteByte('\n')
		for _, sym := range tree.Symbols {
			sb.WriteString(strings.Repeat("\t", level+1))
			sb.WriteByte('(')
			sb.WriteString(fmt.Sprintf("%s %s [%d,%d]", sym.Kind().String(), sym.Name(), sym.Location().Line, sym.Location().Column))

			if childSym, ok := sym.(lib.IChildrenSymbol); ok && childSym.Children() != nil {
				sb.WriteByte('\n')
				treeSexprBuilder(childSym.Children(), sb, level+2)
			}

			sb.WriteByte(')')
		}
	}
	sb.WriteByte(')')
}
