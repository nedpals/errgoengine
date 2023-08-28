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
