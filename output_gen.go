package errgoengine

import (
	"fmt"
	"math"
	"strings"
)

type OutputGenerator struct {
	IsTesting bool
	wr        *strings.Builder
}

func (gen *OutputGenerator) heading(level int, text string) {
	// dont go below zero, dont go above 6
	level = max(min(6, level), 0)
	for i := 0; i < level; i++ {
		gen.wr.WriteByte('#')
	}
	gen.wr.WriteByte(' ')
	gen.writeln(text)
}

func (gen *OutputGenerator) _break() {
	gen.wr.WriteByte('\n')
}

func (gen *OutputGenerator) generateFromExp(level int, explain *ExplainGenerator) {
	if explain.mainExp != nil {
		gen.write(explain.mainExp.String())
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

func (gen *OutputGenerator) writeln(str string, d ...any) {
	if len(str) == 0 {
		return
	}
	gen.write(str, d...)
	gen._break()
}

func (gen *OutputGenerator) write(str string, d ...any) {
	final := fmt.Sprintf(str, d...)
	for _, c := range final {
		if c == '\t' {
			// 1 tab = 4 spaces
			gen.wr.WriteString("    ")
		} else {
			gen.wr.WriteRune(c)
		}
	}
}

func (gen *OutputGenerator) writeLines(lines ...string) {
	for _, line := range lines {
		gen.writeln(line)
	}
}

func (gen *OutputGenerator) Generate(cd *ContextData, explain *ExplainGenerator, bugFix *BugFixGenerator) string {
	if gen.wr == nil {
		gen.wr = &strings.Builder{}
	}

	if len(explain.errorName) != 0 {
		gen.heading(1, explain.errorName)
	}

	gen.generateFromExp(1, explain)
	doc := cd.MainError.Document.CopyContentsOnly()

	if gen.IsTesting {
		startRow := cd.MainError.Nearest.StartPoint().Row
		if startRow-1 == math.MaxUint32 {
			startRow = uint32(cd.MainError.ErrorNode.StartPos.Line)
		}

		startLines := doc.LinesAt(int(startRow)-1, int(startRow)+1)
		endLines := doc.LinesAt(min(int(startRow)+1, doc.TotalLines()), doc.TotalLines())
		arrowLength := int(cd.MainError.Nearest.EndPoint().Row - cd.MainError.Nearest.StartPoint().Row)
		if arrowLength == 0 {
			arrowLength = 1
		}

		startArrowPos := cd.MainError.Nearest.EndPosition().Column
		gen.writeln("```")
		gen.writeLines(startLines...)
		for i := 0; i < startArrowPos-1; i++ {
			if startLines[1][i] == '\t' {
				gen.wr.WriteString("    ")
			} else {
				gen.wr.WriteByte(' ')
			}
		}
		for i := 0; i < arrowLength; i++ {
			gen.wr.WriteByte('^')
		}
		gen._break()
		gen.writeLines(endLines...)
		gen.writeln("```")
	}

	gen.heading(2, "Steps to fix")

	if bugFix.Suggestions != nil && len(bugFix.Suggestions) != 0 {
		for sIdx, s := range bugFix.Suggestions {
			if len(bugFix.Suggestions) == 1 {
				gen.heading(3, s.Title)
			} else {
				gen.heading(3, fmt.Sprintf("%d. %s", sIdx+1, s.Title))
			}

			for idx, step := range s.Steps {
				if len(s.Steps) == 1 {
					gen.writeln(step.Content)
				} else {
					gen.writeln(fmt.Sprintf("%d. %s", idx+1, step.Content))
				}

				if step.Fixes == nil && len(step.Fixes) == 0 {
					continue
				}

				for fIdx, fix := range step.Fixes {
					gen.writeln("```diff")
					startLine := fix.StartPosition.Line
					gen.writeLines(doc.LinesAt(startLine-2, startLine)...)

					gen.write("- ")
					gen.writeln(doc.LineAt(fix.StartPosition.Line))

					gen.write("+ ")
					gen.write(doc.LineAt(fix.StartPosition.Line)[:fix.StartPosition.Column])
					gen.write(fix.NewText)

					if fix.Replace {
						gen.write(doc.LineAt(fix.StartPosition.Line)[fix.EndPosition.Column:])
					}

					gen._break()
					afterLine := startLine
					if fix.Replace {
						afterLine++
					}

					gen.writeLines(doc.LinesAt(afterLine, min(afterLine+2, doc.TotalLines()))...)
					gen.writeln("```")

					if fIdx < len(step.Fixes)-1 {
						gen.writeln(fix.Description)
					} else {
						gen.write(fix.Description)
					}
				}
			}
		}
		// gen._break()
	} else {
		gen.writeln("Nothing to fix")
	}

	return gen.wr.String()
}

func (gen *OutputGenerator) Reset() {
	gen.wr.Reset()
}
