package errgoengine

import (
	"fmt"
	"io/fs"
)

type ExternSymbol struct {
	Name       string         `json:"name"`
	Type       string         `json:"type"`
	ReturnType string         `json:"returnType"`
	Paremeters []ExternSymbol `json:"parameters"`
}

type ExternFile struct {
	Name         string         `json:"name"`
	Package      string         `json:"package"`
	Constructors []ExternSymbol `json:"constructors"`
	Methods      []ExternSymbol `json:"methods"`
}

func ImportExternSymbols(externFs fs.ReadFileFS) (map[string]*SymbolTree, error) {
	if externFs == nil {
		return nil, nil
	}

	symbols := make(map[string]*SymbolTree)
	matches, err := fs.Glob(externFs, "**/*.json")
	if err != nil {
		return nil, err
	}

	for _, match := range matches {
		if err := compileExternSymbol(externFs, match); err != nil {
			return symbols, err
		}
	}

	return symbols, nil
}

func compileExternSymbol(externFs fs.FS, path string) error {
	if externFs == nil {
		return fmt.Errorf("externFs must not be nil")
	}

	return nil
}
