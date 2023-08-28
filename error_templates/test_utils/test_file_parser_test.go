package testutils

import (
	"strings"
	"testing"
)

func TestParserScan(t *testing.T) {
	p := NewParser()
	input := strings.TrimSpace(`
template: "test"
language: "abc"
---
function helloWorld() {}
	`)

	out, err := p.Parse(input)
	if err != nil {
		t.Fatal(err)
	}

	Equals(t, out.Template, "test")
	Equals(t, out.Language, "abc")
	Equals(t, out.Output, strings.TrimSpace(`function helloWorld() {}`))
}

func TestParserInputExpected(t *testing.T) {
	p := NewParser()

	errInput := strings.TrimSpace(`
template: "test"
language: "abc"
---
function helloWorld() {}
	`)

	_, _, err := p.ParseInputExpected("", errInput)
	ExpectError(t, err, "expected 2 raw outputs (1 for input, 1 for expected), got 1")

	errInput = strings.TrimSpace(`
template: "test"
language: "abc"
---
function helloWorld() {}
===
template: "test"
language: "def"
---
eqweqwe
===
template: "test"
language: "ghi"
---
test2
	`)

	_, _, err = p.ParseInputExpected("", errInput)
	ExpectError(t, err, "expected 2 raw outputs (1 for input, 1 for expected), got 3")

	_, _, err = p.ParseInputExpected("", "")
	ExpectError(t, err, "expected 2 raw outputs (1 for input, 1 for expected), got 0")

	input := strings.TrimSpace(`
template: "test"
language: "abc"
---
function helloWorld() {}
===
template: "test2"
language: "def"
---
Error text goes here!
	`)

	inp, exp, err := p.ParseInputExpected("", input)
	if err != nil {
		t.Fatal(err)
	}

	// Input
	Equals(t, inp.Template, "test")
	Equals(t, inp.Language, "abc")
	Equals(t, inp.Output, strings.TrimSpace(`function helloWorld() {}`))

	// Expected
	Equals(t, exp.Template, "test2")
	Equals(t, exp.Language, "def")
	Equals(t, exp.Output, strings.TrimSpace(`Error text goes here!`))
}
