package errgoengine

import (
	"fmt"
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
		if len(line) == 0 {
			gen._break()
		} else {
			gen.writeln(line)
		}
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
	doc := cd.MainError.Document

	if gen.IsTesting {
		startLineNr := cd.MainError.Nearest.StartPosition().Line
		startLines := doc.LinesAt(max(startLineNr-1, 0), startLineNr)
		endLines := doc.LinesAt(min(startLineNr+1, doc.TotalLines()), min(startLineNr+2, doc.TotalLines()))
		arrowLength := int(cd.MainError.Nearest.EndByte() - cd.MainError.Nearest.StartByte())
		if arrowLength == 0 {
			arrowLength = 1
		}

		startArrowPos := cd.MainError.Nearest.StartPosition().Column
		gen.writeln("```")
		gen.writeLines(startLines...)
		for i := 0; i < startArrowPos; i++ {
			if startLines[len(startLines)-1][i] == '\t' {
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
			diffPosition := Position{}
			editedDoc := doc.Editable()

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

				if len(step.Fixes) != 0 {
					descriptionBuilder := &strings.Builder{}

					// get the start and end line after applying the diff
					startLine := step.Fixes[0].StartPosition.Line
					afterLine := step.Fixes[0].EndPosition.Line

					// get the original start and end line
					origStartLine := step.Fixes[0].StartPosition.Line
					origAfterLine := step.Fixes[0].EndPosition.Line

					for fIdx, fix := range step.Fixes {
						diffPosition = diffPosition.addNoCheck(editedDoc.Apply(Changeset{
							NewText:  fix.NewText,
							StartPos: fix.StartPosition,
							EndPos:   fix.EndPosition,
						}.Add(diffPosition)))

						origStartLine = min(origStartLine, fix.StartPosition.Line)
						origAfterLine = max(origAfterLine, fix.EndPosition.Line)

						startLine = min(startLine, fix.StartPosition.Line+diffPosition.Line)
						afterLine = max(afterLine, fix.EndPosition.Line+diffPosition.Line)

						if len(fix.Description) != 0 {
							if fIdx < len(step.Fixes)-1 {
								descriptionBuilder.WriteString(fix.Description + "\n")
							} else {
								descriptionBuilder.WriteString(fix.Description)
							}
						}
					}

					gen.writeln("```diff")
					if startLine > 0 {
						gen.writeLines(editedDoc.LinesAt(startLine-2, startLine-1)...)
					}

					original := editedDoc.LinesAt(origStartLine, origAfterLine)
					for _, origLine := range original {
						gen.write("- ")
						gen.writeln(origLine)
					}

					modified := editedDoc.ModifiedLinesAt(startLine, afterLine)
					for _, modifiedLine := range modified {
						gen.write("+ ")
						gen.writeln(modifiedLine)
					}

					gen.writeLines(editedDoc.LinesAt(origAfterLine+1, min(origAfterLine+2, editedDoc.TotalLines()))...)
					gen.writeln("```")
					if descriptionBuilder.Len() != 0 {
						gen.writeln(descriptionBuilder.String())
					}
				}
			}

			if sIdx < len(bugFix.Suggestions)-1 {
				gen._break()
			}

			editedDoc.Reset()
		}
	} else {
		gen.writeln("Nothing to fix")
	}

	return strings.TrimSpace(gen.wr.String())
}

func (gen *OutputGenerator) Reset() {
	gen.wr.Reset()
}
