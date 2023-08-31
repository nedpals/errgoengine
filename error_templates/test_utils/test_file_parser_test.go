package testutils_test

import (
	"strings"
	"testing"

	tu "github.com/nedpals/errgoengine/error_templates/test_utils"
)

func TestParserScan(t *testing.T) {
	p := tu.NewParser()
	input := strings.TrimSpace(`
template: "abc.test"
---
function helloWorld() {}
	`)

	out, err := p.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}

	tu.Equals(t, out.Template, "test")
	tu.Equals(t, out.Language, "abc")
	tu.Equals(t, out.Output, strings.TrimSpace(`function helloWorld() {}`))
}

func TestParserInputExpected(t *testing.T) {
	p := tu.NewParser()

	errInput := strings.TrimSpace(`
template: "abc.test"
---
function helloWorld() {}
	`)

	_, _, err := p.ParseInputExpected("", errInput)
	tu.ExpectError(t, err, "expected 2 raw outputs (1 for input, 1 for expected), got 1")

	errInput = strings.TrimSpace(`
template: "test"
language: "abc"
---
function helloWorld() {}
===
template: "def.test"
---
eqweqwe
===
template: "ghi.test"
---
test2
	`)

	_, _, err = p.ParseInputExpected("", errInput)
	tu.ExpectError(t, err, "expected 2 raw outputs (1 for input, 1 for expected), got 3")

	_, _, err = p.ParseInputExpected("", "")
	tu.ExpectError(t, err, "expected 2 raw outputs (1 for input, 1 for expected), got 0")

	input := strings.TrimSpace(`
template: "abc.test"
---
function helloWorld() {}
===
template: "def.test2"
---
Error text goes here!
	`)

	inp, exp, err := p.ParseInputExpected("", input)
	if err != nil {
		t.Fatal(err)
	}

	// Input
	tu.Equals(t, inp.Template, "test")
	tu.Equals(t, inp.Language, "abc")
	tu.Equals(t, inp.Output, strings.TrimSpace(`function helloWorld() {}`))

	// Expected
	tu.Equals(t, exp.Template, "test2")
	tu.Equals(t, exp.Language, "def")
	tu.Equals(t, exp.Output, strings.TrimSpace(`Error text goes here!`))
}
