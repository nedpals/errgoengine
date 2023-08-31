package testutils_test

import (
	"strings"
	"testing"

	etu "github.com/nedpals/errgoengine/error_templates/test_utils"
	testutils "github.com/nedpals/errgoengine/test_utils"
)

func TestParserScan(t *testing.T) {
	p := etu.NewParser()
	input := strings.TrimSpace(`
template: "abc.test"
---
function helloWorld() {}
	`)

	out, err := p.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}

	testutils.Equals(t, out.Template, "test")
	testutils.Equals(t, out.Language, "abc")
	testutils.Equals(t, out.Output, strings.TrimSpace(`function helloWorld() {}`))
}

func TestParserInputExpected(t *testing.T) {
	p := etu.NewParser()

	errInput := strings.TrimSpace(`
template: "abc.test"
---
function helloWorld() {}
	`)

	_, _, err := p.ParseInputExpected("", errInput)
	testutils.ExpectError(t, err, "expected 2 raw outputs (1 for input, 1 for expected), got 1")

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
	testutils.ExpectError(t, err, "expected 2 raw outputs (1 for input, 1 for expected), got 3")

	_, _, err = p.ParseInputExpected("", "")
	testutils.ExpectError(t, err, "expected 2 raw outputs (1 for input, 1 for expected), got 0")

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
	testutils.Equals(t, inp.Template, "test")
	testutils.Equals(t, inp.Language, "abc")
	testutils.Equals(t, inp.Output, strings.TrimSpace(`function helloWorld() {}`))

	// Expected
	testutils.Equals(t, exp.Template, "test2")
	testutils.Equals(t, exp.Language, "def")
	testutils.Equals(t, exp.Output, strings.TrimSpace(`Error text goes here!`))
}
