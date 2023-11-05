package errgoengine

import (
	"fmt"
	"math"
	"strings"
)

type OutputGenerator interface {
	Generate(*ContextData, *ExplainGenerator, *BugFixGenerator) string
	Reset()
}

type MarkdownOutputGenerator struct {
	IsTesting bool
	wr        *strings.Builder
}

func (gen *MarkdownOutputGenerator) heading(level int, text string) {
	// dont go below zero, dont go above 6
	level = max(min(6, level), 0)
	for i := 0; i < level; i++ {
		gen.wr.WriteByte('#')
	}
	gen.wr.WriteByte(' ')
	gen.wr.WriteString(text)
	gen._break()
}

func (gen *MarkdownOutputGenerator) _break() {
	gen.wr.WriteByte('\n')
}

func (gen *MarkdownOutputGenerator) generateFromExp(level int, explain *ExplainGenerator) {
	if explain.mainExp != nil {
		gen.wr.WriteString(explain.mainExp.String())
	}

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

	if len(explain.errorName) != 0 {
		gen.heading(1, explain.errorName)
	}

	gen.generateFromExp(1, explain)

	lines := strings.Split(cd.MainError.Document.Contents, "\n")

	if gen.IsTesting {
		startRow := cd.MainError.Nearest.StartPoint().Row
		if startRow-1 == math.MaxUint32 {
			startRow = uint32(cd.MainError.ErrorNode.Line)
		}

		startLines := lines[max(startRow-1, 0) : startRow+1]
		endLines := lines[min(int(startRow)+1, len(lines)):]

		arrowLength := int(cd.MainError.Nearest.EndPoint().Row - cd.MainError.Nearest.StartPoint().Row)
		if arrowLength == 0 {
			arrowLength = 1
		}

		startArrowPos := int(cd.MainError.Nearest.StartPoint().Column)

		gen.wr.WriteString("```\n")
		for _, line := range startLines {
			gen.wr.WriteString(line)
			gen.wr.WriteByte('\n')
		}
		for i := 0; i < startArrowPos; i++ {
			gen.wr.WriteByte(' ')
		}
		for i := 0; i < arrowLength; i++ {
			gen.wr.WriteByte('^')
		}
		gen.wr.WriteByte('\n')
		for _, line := range endLines {
			gen.wr.WriteString(line)
			gen.wr.WriteByte('\n')
		}
		gen.wr.WriteString("```\n")
	}

	gen.heading(2, "Steps to fix")

	if bugFix.Steps != nil && len(bugFix.Steps) != 0 {
		for idx, step := range bugFix.Steps {
			gen.wr.WriteString(fmt.Sprintf("%d. %s\n", idx+1, step.Explanation))
			if step.Fixes != nil && len(step.Fixes) != 0 {
				gen.wr.WriteString("```diff\n")
				for _, fix := range step.Fixes {
					startLine := fix.Position.Line
					for i := startLine - 2; i < startLine; i++ {
						gen.wr.WriteString(lines[i])
						gen.wr.WriteByte('\n')
					}

					gen.wr.WriteString("+ ")
					for i := 0; i < fix.Position.Column; i++ {
						gen.wr.WriteByte(' ')
					}
					gen.wr.WriteString(fix.NewText)
					gen.wr.WriteByte('\n')

					if fix.Replace {
						gen.wr.WriteString("- ")
						gen.wr.WriteString(lines[fix.Position.Line])
						gen.wr.WriteByte('\n')
					}

					afterLine := startLine
					if fix.Replace {
						afterLine++
					}

					for i := afterLine; i < min(afterLine+2, len(lines)); i++ {
						gen.wr.WriteString(lines[i])
						gen.wr.WriteByte('\n')
					}
				}
				gen.wr.WriteString("\n```")
			}
		}
	} else {
		gen.wr.WriteString("Nothing to fix")
	}

	return gen.wr.String()
}

func (gen *MarkdownOutputGenerator) Reset() {
	gen.wr.Reset()
}
