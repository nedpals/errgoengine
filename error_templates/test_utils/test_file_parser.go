package testutils

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"text/scanner"
)

var importantKeys = []string{"language", "template", "output"}

type TestOutput struct {
	Name     string
	Language string
	Template string
	Output   string
}

type Parser struct {
	prevTok token
	tok     token
	nextTok token
	sc      *scanner.Scanner
}

func NewParser() *Parser {
	return &Parser{
		sc: &scanner.Scanner{},
	}
}

type token struct {
	tok  rune
	text string
	pos  scanner.Position
}

func debugf(format string, a ...any) {
	// if !IsTestRun() {
	// 	return
	// }
	// fmt.Printf(format+"\n", a...)
}

func (t token) isEOF() bool {
	return t.tok == scanner.EOF
}

func (t token) areAdjacentWith(t2 token) bool {
	// fmt.Printf("[areAdjacentWith] t len: %d, offset: %d\n", t.len(), t2.pos.Offset)
	return t.pos.Offset == t2.len()
}

func (t token) len() int {
	return t.pos.Offset + len(t.text)
}

func (t token) kind() string {
	return stripKind(scanner.TokenString(t.tok))
}

func stripKind(kind string) string {
	if len(kind) == 3 && strings.HasPrefix(kind, `"`) && strings.HasSuffix(kind, `"`) {
		return strings.TrimSuffix(strings.TrimPrefix(kind, `"`), `"`)
	}
	return kind
}

func (p *Parser) token(tok rune) token {
	return token{
		tok:  tok,
		text: p.sc.TokenText(),
		pos:  p.sc.Position,
	}
}

func (p *Parser) Scan() token {
	p.prevTok = p.tok
	debugf("prevTok = `%s`", p.prevTok.kind())
	if p.tok.len() == 0 {
		newTokRune := p.sc.Scan()
		p.tok = p.token(newTokRune)
	} else {
		p.tok = p.nextTok
	}
	debugf("tok = `%s`", p.tok.kind())
	newNextTokRune := p.sc.Scan()
	p.nextTok = p.token(newNextTokRune)
	debugf("nextTok = `%s`", p.nextTok.kind())
	return p.tok
}

func parserError(tok token, msg string) error {
	return fmt.Errorf("%s: %s", tok.pos, msg)
}

func expectError(gotTok token, exp string) error {
	got := gotTok.kind()
	isChar := len(got) == 3 && strings.HasPrefix(got, "\"") && strings.HasSuffix(got, "\"")
	text := ""
	if !isChar {
		text = " (" + gotTok.text + ")"
	}
	return parserError(gotTok, fmt.Sprintf("expected `%s`, got `%s`%s", exp, got, text))
}

func (p *Parser) expect(exp string) error {
	debugf("[Parser.expect] asserting tok is `%s` (tok kind = `%s`)", exp, p.nextTok.kind())
	if got := p.nextTok.kind(); got != exp {
		return expectError(p.nextTok, exp)
	}
	p.Scan()
	return nil
}

func (p *Parser) expectNextTo(exp string) error {
	if err := p.expect(exp); err != nil {
		return err
	}
	if !p.tok.areAdjacentWith(p.prevTok) {
		return parserError(
			p.tok,
			fmt.Sprintf(
				"`%s` should be adjacent to `%s` (exp %d, got %d)",
				p.tok.text,
				p.prevTok.text,
				p.prevTok.len(),
				p.tok.pos.Offset),
		)
	}
	return nil
}

func (p *Parser) Parse(text string) (*TestOutput, error) {
	// reset tok states first
	p.prevTok = token{}
	p.tok = token{}
	p.nextTok = token{}

	// resetting scanner instance with new text
	p.sc.Init(strings.NewReader(text))

	kv := map[string]string{
		"name":     "",
		"language": "",
		"template": "",
		"output":   "",
	}

	for !p.tok.isEOF() {
		p.Scan()
		kind := p.tok.kind()
		debugf("====> RECEIVED TOK: " + p.tok.kind())

		if kind == "Ident" {
			key := p.tok.text
			if err := p.expectNextTo(":"); err != nil {
				return nil, err
			}
			if err := p.expect("String"); err != nil {
				return nil, err
			}

			value, err := strconv.Unquote(p.tok.text)
			if err != nil {
				return nil, err
			}

			kv[key] = value
		} else if kind == "=" {
			firstSepCol := p.tok.pos.Column
			if err := p.expectNextTo("="); err != nil {
				return nil, err
			}

			// ===
			if p.tok.text != p.nextTok.text {
				return nil, expectError(p.tok, "=")
			}

			if firstSepCol != 1 || p.sc.Peek() != '\n' {
				return nil, parserError(p.tok, "output separator should be right after line break and should have no trailing whitespaces")
			}

			p.sc.Next()
			buf := new(bytes.Buffer)
			for t := p.sc.Next(); t != scanner.EOF; t = p.sc.Next() {
				debugf("next = %s", strconv.QuoteRune(t))
				buf.WriteRune(t)
			}

			kv["output"] = buf.String()

			// scan to jump to EOF
			p.Scan()
		}
	}

	// check if all entries are present
	for _, k := range importantKeys {
		if v, ok := kv[k]; !ok {
			return nil, fmt.Errorf("missing %s", k)
		} else if len(v) == 0 {
			return nil, fmt.Errorf("missing %s", k)
		}
	}

	return &TestOutput{
		Name:     kv["name"],
		Language: kv["language"],
		Template: kv["template"],
		Output:   kv["output"],
	}, nil
}
