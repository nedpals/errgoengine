package errgoengine

import (
	"fmt"
	"strings"
)

type OutputGenerator interface {
	Generate(*ContextData, *ExplainGenerator, *BugFixGenerator) string
}

type MarkdownOutputGenerator struct {
	errorName string
	wr        *strings.Builder
}

func (gen *MarkdownOutputGenerator) heading(level int, text string) {
	if level > 6 {
		level = 6
	} else if level <= 0 {
		level = 1
	}
	for i := 0; i < level; i++ {
		gen.wr.WriteByte('#')
	}
	gen.wr.WriteByte(' ')
	gen.wr.WriteString(text)
	gen._break()
}

func (gen *MarkdownOutputGenerator) _break() {
	gen.wr.WriteByte('\n')
	gen.wr.WriteByte('\n')
}

func (gen *MarkdownOutputGenerator) generateFromExp(level int, explain *ExplainGenerator) {
	gen.wr.WriteString(explain.mainExp.String())
	if explain.sections != nil {
		for sectionName, exp := range explain.sections {
			gen._break()
			gen.heading(level+1, sectionName)
			gen.generateFromExp(level+1, exp)
		}
	} else {
		gen._break()
	}
}

func (gen *MarkdownOutputGenerator) Generate(cd *ContextData, explain *ExplainGenerator, bugFix *BugFixGenerator) string {
	if gen.wr == nil {
		gen.wr = &strings.Builder{}
	}

	if len(gen.errorName) != 0 {
		gen.heading(1, gen.errorName)
	}

	gen.generateFromExp(1, explain)
	gen.heading(2, "Steps to fix")

	if bugFix.Fixes != nil && len(bugFix.Fixes) != 0 {
		for idx, fix := range bugFix.Fixes {
			gen.wr.WriteString(fmt.Sprintf("%d. %s\n", idx+1, fix.Explanation))

			// TODO: generate syntax highlight
		}
	} else {
		gen.wr.WriteString("Nothing to fix")
	}

	return gen.wr.String()
}
