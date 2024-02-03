package errgoengine

import (
	"fmt"
	"strings"
)

type OutputGenerator struct {
	IsTesting bool
	Builder   *strings.Builder
}

func (gen *OutputGenerator) Heading(level int, text string) {
	// dont go below zero, dont go above 6
	level = max(min(6, level), 0)
	for i := 0; i < level; i++ {
		gen.Builder.WriteByte('#')
	}
	gen.Builder.WriteByte(' ')
	gen.Writeln(text)
}

func (gen *OutputGenerator) _break() {
	gen.Builder.WriteByte('\n')
}

func (gen *OutputGenerator) ExpGen(level int, explain *ExplainGenerator) {
	if explain.Builder != nil {
		gen.Write(explain.Builder.String())
	}

	if explain.Sections != nil {
		for sectionName, exp := range explain.Sections {
			gen._break()
			gen.Heading(level+1, sectionName)
			gen.ExpGen(level+1, exp)
		}
	} else {
		gen._break()
	}
}

func (gen *OutputGenerator) Writeln(str string, d ...any) {
	if len(str) == 0 {
		return
	}
	gen.Write(str, d...)
	gen._break()
}

func (gen *OutputGenerator) Write(str string, d ...any) {
	final := fmt.Sprintf(str, d...)
	for _, c := range final {
		if c == '\t' {
			// 1 tab = 4 spaces
			gen.Builder.WriteString("    ")
		} else {
			gen.Builder.WriteRune(c)
		}
	}
}

func (gen *OutputGenerator) WriteLines(lines ...string) {
	for _, line := range lines {
		if len(line) == 0 {
			gen._break()
		} else {
			gen.Writeln(line)
		}
	}
}

func (gen *OutputGenerator) Generate(cd *ContextData, explain *ExplainGenerator, bugFix *BugFixGenerator) string {
	if gen.Builder == nil {
		gen.Builder = &strings.Builder{}
	}

	if len(explain.ErrorName) != 0 {
		gen.Heading(1, explain.ErrorName)
	}

	gen.ExpGen(1, explain)
	doc := cd.MainError.Document

	if doc != nil && gen.IsTesting && !cd.MainError.Nearest.IsNull() {
		startLineNr := cd.MainError.Nearest.StartPosition().Line
		startLines := doc.LinesAt(max(startLineNr-1, 0), startLineNr)
		endLines := doc.LinesAt(min(startLineNr+1, doc.TotalLines()), min(startLineNr+2, doc.TotalLines()))
		arrowLength := int(cd.MainError.Nearest.EndByte() - cd.MainError.Nearest.StartByte())
		if arrowLength == 0 {
			arrowLength = 1
		}

		startArrowPos := cd.MainError.Nearest.StartPosition().Column
		gen.Writeln("```")
		gen.WriteLines(startLines...)
		for i := 0; i < startArrowPos; i++ {
			if startLines[len(startLines)-1][i] == '\t' {
				gen.Builder.WriteString("    ")
			} else {
				gen.Builder.WriteByte(' ')
			}
		}
		for i := 0; i < arrowLength; i++ {
			gen.Builder.WriteByte('^')
		}
		gen._break()
		gen.WriteLines(endLines...)
		gen.Writeln("```")
	}

	gen.Heading(2, "Steps to fix")

	if bugFix.Suggestions != nil && len(bugFix.Suggestions) != 0 {
		for sIdx, s := range bugFix.Suggestions {
			if len(bugFix.Suggestions) == 1 {
				gen.Heading(3, s.Title)
			} else {
				gen.Heading(3, fmt.Sprintf("%d. %s", sIdx+1, s.Title))
			}

			for idx, step := range s.Steps {
				if len(s.Steps) == 1 {
					gen.Writeln(step.Content)
				} else {
					gen.Writeln(fmt.Sprintf("%d. %s", idx+1, step.Content))
				}

				if step.Fixes == nil && len(step.Fixes) == 0 {
					continue
				}

				if len(step.Fixes) != 0 {
					descriptionBuilder := &strings.Builder{}

					// get the start and end line after applying the diff
					startLine := step.StartLine
					afterLine := step.AfterLine

					// get the original start and end line
					origStartLine := step.OrigStartLine
					origAfterLine := step.OrigAfterLine

					gen.Writeln("```diff")

					// use origStartLine instead of startLine because we want to show the original lines
					if startLine > 0 {
						deduct := -2
						if step.DiffPosition.Line < 0 {
							deduct += step.DiffPosition.Line
						}
						gen.WriteLines(step.Doc.LinesAt(origStartLine+deduct, origStartLine-1)...)
					}

					modified := step.Doc.ModifiedLinesAt(startLine, afterLine)
					original := step.Doc.LinesAt(origStartLine, origAfterLine)
					for i, origLine := range original {
						if i >= len(modified) || modified[i] != origLine {
							gen.Write("- ")
						}
						if len(origLine) == 0 {
							gen._break()
						} else {
							gen.Writeln(origLine)
						}
					}

					// show this only if the total is not negative
					if startLine >= origStartLine && afterLine >= origAfterLine {
						// TODO: redundant
						modified := step.Doc.ModifiedLinesAt(startLine, afterLine)
						// TODO: merge with previous `original` variable
						originalLines := doc.LinesAt(origStartLine, min(origAfterLine+step.DiffPosition.Line, doc.TotalLines()))
						for i, modifiedLine := range modified {
							if i == 0 && len(modified) == 1 && len(modifiedLine) == 0 {
								continue
							}
							// skip marking as "addition" if the lines are the same
							if i < len(originalLines) && modifiedLine == originalLines[i] {
								// write only if the line is not the last line
								if startLine+i < origAfterLine {
									gen.Write(modifiedLine)
									gen._break()
								}
								continue
							}
							gen.Write("+")
							if len(modifiedLine) != 0 {
								gen.Write(" ")
							}
							gen.Write(modifiedLine)
							gen._break()
						}
					}

					gen.WriteLines(step.Doc.LinesAt(origAfterLine+1, min(origAfterLine+2, step.Doc.TotalLines()))...)
					gen.Writeln("```")

					for fIdx, fix := range step.Fixes {
						if len(fix.Description) != 0 {
							if fIdx < len(step.Fixes)-1 {
								descriptionBuilder.WriteString(fix.Description + "\n")
							} else {
								descriptionBuilder.WriteString(fix.Description)
							}
						}
					}

					if descriptionBuilder.Len() != 0 {
						gen.Writeln(descriptionBuilder.String())
					}
				}
			}

			if sIdx < len(bugFix.Suggestions)-1 {
				gen._break()
			}

		}
	} else {
		gen.Writeln("Nothing to fix")
	}

	return strings.TrimSpace(gen.Builder.String())
}

func (gen *OutputGenerator) Reset() {
	gen.Builder.Reset()
}
