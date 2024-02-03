package errgoengine_test

import (
	"strings"
	"testing"

	lib "github.com/nedpals/errgoengine"
	sitter "github.com/smacker/go-tree-sitter"
)

func TestExplainGenerator(t *testing.T) {
	t.Run("errorName", func(t *testing.T) {
		gen := &lib.ExplainGenerator{ErrorName: "TestError"}

		if gen.ErrorName != "TestError" {
			t.Errorf("Expected 'TestError', got %s", gen.ErrorName)
		}
	})

	t.Run("Add", func(t *testing.T) {
		t.Run("Simple", func(t *testing.T) {
			gen := &lib.ExplainGenerator{}
			gen.Add("This is a simple error explanation.")

			if gen.Builder.String() != "This is a simple error explanation." {
				t.Errorf("Expected 'This is a simple error explanation.', got %s", gen.Builder.String())
			}
		})

		t.Run("Simple with string data", func(t *testing.T) {
			gen := &lib.ExplainGenerator{}
			gen.Add("This is a simple error explanation with data: %s", "Hello")

			if gen.Builder.String() != "This is a simple error explanation with data: Hello" {
				t.Errorf("Expected 'This is a simple error explanation with data: Hello', got %s", gen.Builder.String())
			}
		})

		t.Run("Simple with int data", func(t *testing.T) {
			gen := &lib.ExplainGenerator{}
			gen.Add("This is a simple error explanation with data: %d", 10)

			if gen.Builder.String() != "This is a simple error explanation with data: 10" {
				t.Errorf("Expected 'This is a simple error explanation with data: 10', got %s", gen.Builder.String())
			}
		})

		t.Run("Simple with mixed data", func(t *testing.T) {
			gen := &lib.ExplainGenerator{}
			gen.Add("This is a simple error explanation with data: %s and %d", "Hello", 10)

			if gen.Builder.String() != "This is a simple error explanation with data: Hello and 10" {
				t.Errorf("Expected 'This is a simple error explanation with data: Hello and 10', got %s", gen.Builder.String())
			}
		})

		t.Run("Append", func(t *testing.T) {
			gen := &lib.ExplainGenerator{}
			gen.Add("This is a simple error explanation.")
			gen.Add("This is another error message.")

			if gen.Builder.String() != "This is a simple error explanation.This is another error message." {
				t.Errorf("Expected 'This is a simple error explanation.This is another error message.', got %s", gen.Builder.String())
			}
		})

		t.Run("Append with newline", func(t *testing.T) {
			gen := &lib.ExplainGenerator{ErrorName: "TestError"}
			gen.Add("This is a simple error explanation.\n")
			gen.Add("This is another error message.")

			if gen.Builder.String() != "This is a simple error explanation.\nThis is another error message." {
				t.Errorf("Expected 'This is a simple error explanation.\nThis is another error message.', got %s", gen.Builder.String())
			}
		})

		t.Run("Append with string data", func(t *testing.T) {
			gen := &lib.ExplainGenerator{}
			gen.Add("This is a simple error explanation with data: %s", "Hello")
			gen.Add("This is another error message with data: %s", "World")

			if gen.Builder.String() != "This is a simple error explanation with data: HelloThis is another error message with data: World" {
				t.Errorf("Expected 'This is a simple error explanation with data: HelloThis is another error message with data: World', got %s", gen.Builder.String())
			}
		})

		t.Run("Append with int data", func(t *testing.T) {
			gen := &lib.ExplainGenerator{}
			gen.Add("This is a simple error explanation with data: %d", 10)
			gen.Add("This is another error message with data: %d", 20)

			if gen.Builder.String() != "This is a simple error explanation with data: 10This is another error message with data: 20" {
				t.Errorf("Expected 'This is a simple error explanation with data: 10This is another error message with data: 20', got %s", gen.Builder.String())
			}
		})

		t.Run("Append with mixed data", func(t *testing.T) {
			gen := &lib.ExplainGenerator{}
			gen.Add("This is a simple error explanation with data: %s", "Hello")
			gen.Add("This is another error message with data: %d", 20)

			if gen.Builder.String() != "This is a simple error explanation with data: HelloThis is another error message with data: 20" {
				t.Errorf("Expected 'This is a simple error explanation with data: HelloThis is another error message with data: 20', got %s", gen.Builder.String())
			}
		})

		t.Run("Empty", func(t *testing.T) {
			gen := &lib.ExplainGenerator{}
			gen.Add("")

			if gen.Builder.String() != "" {
				t.Errorf("Expected '', got %s", gen.Builder.String())
			}
		})
	})

	t.Run("CreateSection", func(t *testing.T) {
		t.Run("Simple", func(t *testing.T) {
			gen := &lib.ExplainGenerator{}

			section := gen.CreateSection("TestSection")

			if gen.Sections == nil {
				t.Errorf("Expected sections to be created")
			}

			if _, ok := gen.Sections["TestSection"]; !ok {
				t.Errorf("Expected section to be added to sections")
			}

			if section == nil {
				t.Errorf("Expected section to be created")
			}
		})

		t.Run("Empty", func(t *testing.T) {
			gen := &lib.ExplainGenerator{}

			section := gen.CreateSection("")

			if _, ok := gen.Sections[""]; ok {
				t.Errorf("Expected section to not be added to sections")
			}

			if section != nil {
				t.Errorf("Expected section to not be created")
			}
		})

		t.Run("Writable section", func(t *testing.T) {
			gen := &lib.ExplainGenerator{}

			section := gen.CreateSection("TestSection")
			section.Add("This is a simple error message.")

			if gen.Sections["TestSection"].Builder.String() != "This is a simple error message." {
				t.Errorf("Expected 'This is a simple error message.', got %s", gen.Sections["TestSection"].Builder.String())
			}
		})
	})
}

func TestBugFixGenerator(t *testing.T) {
	parser := sitter.NewParser()
	// create a parsed document
	doc, err := lib.ParseDocument("hello.test", strings.NewReader("print('Hello, World!')"), parser, lib.TestLanguage, nil)
	if err != nil {
		t.Errorf("Error parsing document: %s", err)
	}

	t.Run("Add", func(t *testing.T) {
		t.Run("Simple", func(t *testing.T) {
			gen := lib.NewBugFixGenerator(doc)

			gen.Add("A descriptive suggestion sentence or phrase", func(s *lib.BugFixSuggestion) {
				if s.Title != "A descriptive suggestion sentence or phrase" {
					t.Errorf("Expected 'A descriptive suggestion sentence or phrase', got %s", s.Title)
				}

				if s.Doc == nil {
					t.Errorf("Expected document to be set")
				}
			})
		})

		t.Run("Empty title", func(t *testing.T) {
			gen := lib.NewBugFixGenerator(doc)

			err := gen.Add("", func(s *lib.BugFixSuggestion) {})
			if err == nil {
				t.Errorf("Expected error, got nil")
			}

			if err.Error() != "title cannot be empty" {
				t.Errorf("Expected 'title cannot be empty', got %s", err.Error())
			}
		})

		t.Run("Empty function", func(t *testing.T) {
			gen := lib.NewBugFixGenerator(doc)

			err := gen.Add("A descriptive suggestion sentence or phrase", nil)
			if err == nil {
				t.Errorf("Expected error, got nil")
			}

			if err.Error() != "maker function cannot be nil" {
				t.Errorf("Expected 'maker function cannot be nil', got %s", err.Error())
			}
		})

		t.Run("Multiple", func(t *testing.T) {
			gen := lib.NewBugFixGenerator(doc)

			gen.Add("A descriptive suggestion sentence or phrase", func(s *lib.BugFixSuggestion) {})
			gen.Add("This is another error message.", func(s *lib.BugFixSuggestion) {})

			if len(gen.Suggestions) != 2 {
				t.Errorf("Expected 2 suggestions, got %d", len(gen.Suggestions))
			}
		})
	})

	t.Run("Suggestion/AddStep", func(t *testing.T) {
		t.Run("Simple", func(t *testing.T) {
			gen := lib.NewBugFixGenerator(doc)

			gen.Add("A descriptive suggestion sentence or phrase", func(s *lib.BugFixSuggestion) {
				s.AddStep("This is a step.", func(step *lib.BugFixStep) {
					if step.Content != "This is a step." {
						t.Errorf("Expected 'This is a step.', got %s", step.Content)
					}

					if step.Doc == nil {
						t.Errorf("Expected document to be set")
					}
				})
			})
		})

		t.Run("Without period", func(t *testing.T) {
			gen := lib.NewBugFixGenerator(doc)

			gen.Add("A descriptive suggestion sentence or phrase", func(s *lib.BugFixSuggestion) {
				s.AddStep("This is a step", func(step *lib.BugFixStep) {
					if step.Content != "This is a step." {
						t.Errorf("Expected 'This is a step.', got %s", step.Content)
					}
				})
			})
		})

		t.Run("With punctuation", func(t *testing.T) {
			gen := lib.NewBugFixGenerator(doc)

			gen.Add("A descriptive suggestion sentence or phrase", func(s *lib.BugFixSuggestion) {
				s.AddStep("Oh wow!", func(step *lib.BugFixStep) {
					if step.Content != "Oh wow!" {
						t.Errorf("Expected 'Oh wow!', got %s", step.Content)
					}
				})
			})
		})

		t.Run("With string data", func(t *testing.T) {
			gen := lib.NewBugFixGenerator(doc)

			gen.Add("A descriptive suggestion sentence or phrase", func(s *lib.BugFixSuggestion) {
				s.AddStep("This is a step with data: %s", "Hello", func(step *lib.BugFixStep) {
					if step.Content != "This is a step with data: Hello." {
						t.Errorf("Expected 'This is a step with data: Hello.', got %s", step.Content)
					}
				})
			})
		})

		t.Run("With int data", func(t *testing.T) {
			gen := lib.NewBugFixGenerator(doc)

			gen.Add("A descriptive suggestion sentence or phrase", func(s *lib.BugFixSuggestion) {
				s.AddStep("This is a step with data: %d", 10, func(step *lib.BugFixStep) {
					if step.Content != "This is a step with data: 10." {
						t.Errorf("Expected 'This is a step with data: 10.', got %s", step.Content)
					}
				})
			})
		})

		t.Run("With mixed data", func(t *testing.T) {
			gen := lib.NewBugFixGenerator(doc)

			gen.Add("A descriptive suggestion sentence or phrase", func(s *lib.BugFixSuggestion) {
				s.AddStep("This is a step with data: %s and %d", "Hello", 10, func(step *lib.BugFixStep) {
					if step.Content != "This is a step with data: Hello and 10." {
						t.Errorf("Expected 'This is a step with data: Hello and 10.', got %s", step.Content)
					}
				})
			})
		})

		t.Run("Empty content", func(t *testing.T) {
			gen := lib.NewBugFixGenerator(doc)

			gen.Add("A descriptive suggestion sentence or phrase", func(s *lib.BugFixSuggestion) {
				// recover
				defer func() {
					if r := recover(); r != nil {
						err := r.(error)
						if err.Error() != "content cannot be empty" {
							t.Errorf("Expected 'content cannot be empty', got %s", err.Error())
						}
					}
				}()

				s.AddStep("", func(step *lib.BugFixStep) {})
			})
		})
	})

	t.Run("Suggestion/AddFix/Add content", func(t *testing.T) {
		gen := &lib.BugFixGenerator{
			Document: doc,
		}

		gen.Add("A descriptive suggestion sentence or phrase", func(s *lib.BugFixSuggestion) {
			// add a = 1 fix
			step := s.AddStep("This is a step.").
				AddFix(lib.FixSuggestion{
					NewText: "\na = 1",
					StartPosition: lib.Position{
						Line:   0,
						Column: 22,
					},
					EndPosition: lib.Position{
						Line:   0,
						Column: 22,
					},
				})

			if step.Doc.String() != "print('Hello, World!')\na = 1" {
				t.Errorf("Expected 'print('Hello, World!')\na = 1', got %s", step.Doc.String())
			}

			if len(s.Steps[0].Fixes) != 1 {
				t.Errorf("Expected 1 fix, got %d", len(s.Steps[0].Fixes))
			}
		})
	})

	t.Run("Suggestion/AddFix/Update", func(t *testing.T) {
		gen := &lib.BugFixGenerator{
			Document: doc,
		}

		gen.Add("A descriptive suggestion sentence or phrase", func(s *lib.BugFixSuggestion) {
			step := s.AddStep("This is a step.").
				AddFix(lib.FixSuggestion{
					NewText: "Welcome to the world",
					StartPosition: lib.Position{
						Line:   0,
						Column: 7,
					},
					EndPosition: lib.Position{
						Line:   0,
						Column: 20,
					},
				})

			if step.Doc.String() != "print('Welcome to the world')" {
				t.Errorf("Expected 'print('Welcome to the world')', got %s", step.Doc.String())
			}

			// add a = 1 fix
			step.AddFix(lib.FixSuggestion{
				NewText: "\na = 1",
				StartPosition: lib.Position{
					Line:   0,
					Column: 22,
				},
				EndPosition: lib.Position{
					Line:   0,
					Column: 22,
				},
			})

			if step.Doc.String() != "print('Welcome to the world')\na = 1" {
				t.Errorf("Expected 'print('Welcome to the world')\na = 1', got %s", step.Doc.String())
			}
		})
	})

	t.Run("Suggestion/AddFix/Delete", func(t *testing.T) {
		gen := &lib.BugFixGenerator{
			Document: doc,
		}

		gen.Add("A descriptive suggestion sentence or phrase", func(s *lib.BugFixSuggestion) {
			// removes the content inside the print function
			step := s.AddStep("This is a step.").
				AddFix(lib.FixSuggestion{
					NewText: "",
					StartPosition: lib.Position{
						Line:   0,
						Column: 7,
					},
					EndPosition: lib.Position{
						Line:   0,
						Column: 20,
					},
				})

			if step.Doc.String() != "print('')" {
				t.Errorf("Expected '', got %s", step.Doc.String())
			}
		})
	})

	t.Run("Suggestion/AddFix/Mixed", func(t *testing.T) {
		gen := &lib.BugFixGenerator{
			Document: doc,
		}

		gen.Add("A descriptive suggestion sentence or phrase", func(s *lib.BugFixSuggestion) {
			// removes the content inside the print function
			step := s.AddStep("This is a step.").
				AddFix(lib.FixSuggestion{
					NewText: "",
					StartPosition: lib.Position{
						Line:   0,
						Column: 7,
					},
					EndPosition: lib.Position{
						Line:   0,
						Column: 20,
					},
				})

			if step.Doc.String() != "print('')" {
				t.Errorf("Expected 'print('')', got %s", step.Doc.String())
			}

			// add a text inside the print function
			step.AddFix(lib.FixSuggestion{
				NewText: "Hello, World!",
				StartPosition: lib.Position{
					Line:   0,
					Column: 7,
				},
				EndPosition: lib.Position{
					Line:   0,
					Column: 7,
				},
			})

			if step.Doc.String() != "print('Hello, World!')" {
				t.Errorf("Expected 'print('Hello, World!')', got %s", step.Doc.String())
			}

			// add a = 1 fix
			step.AddFix(lib.FixSuggestion{
				NewText: "\na = 1",
				StartPosition: lib.Position{
					Line:   0,
					Column: 24,
				},
				EndPosition: lib.Position{
					Line:   0,
					Column: 24,
				},
			})

			if step.Doc.String() != "print('Hello, World!')\na = 1" {
				t.Errorf("Expected 'print('Hello, World!')\na = 1', got %s", step.Doc.String())
			}
		})
	})

	t.Run("Suggestion/MultipleSteps/AddFix", func(t *testing.T) {
		gen := &lib.BugFixGenerator{
			Document: doc,
		}

		gen.Add("Improve the print call", func(s *lib.BugFixSuggestion) {
			step := s.AddStep("Remove the content inside the print function").
				AddFix(lib.FixSuggestion{
					NewText: "",
					StartPosition: lib.Position{
						Line:   0,
						Column: 7,
					},
					EndPosition: lib.Position{
						Line:   0,
						Column: 20,
					},
				})

			if step.Doc.String() != "print('')" {
				t.Errorf("Expected 'print('')', got %s", step.Doc.String())
			}

			// add a text inside the print function
			step2 := s.AddStep("Add a custom text").
				AddFix(lib.FixSuggestion{
					NewText: "Foo bar?",
					StartPosition: lib.Position{
						Line:   0,
						Column: 7,
					},
					EndPosition: lib.Position{
						Line:   0,
						Column: 7,
					},
				})

			if step2.Doc.String() != "print('Foo bar?')" {
				t.Errorf("Expected 'print('Foo bar?')', got %s", step2.Doc.String())
			}

			step3 := s.AddStep("Add an assignment below").AddFix(lib.FixSuggestion{
				NewText: "\nx = 2",
				StartPosition: lib.Position{
					Line:   0,
					Column: 24,
				},
				EndPosition: lib.Position{
					Line:   0,
					Column: 24,
				},
			})

			if step3.Doc.String() != "print('Foo bar?')\nx = 2" {
				t.Errorf("Expected 'print('Foo bar?')\nx = 2', got %s", step3.Doc.String())
			}
		})
	})
}
