package java

import (
	"fmt"

	lib "github.com/nedpals/errgoengine"
)

type javaBuiltinTypeStore struct {
	typesSymbols map[string]lib.Symbol
}

func (store *javaBuiltinTypeStore) Builtin(name string) lib.Symbol {
	if store.typesSymbols == nil {
		store.typesSymbols = make(map[string]lib.Symbol)
	} else if sym, ok := store.FindByName(name); ok {
		return sym
	}
	store.typesSymbols[name] = lib.Builtin(name)
	return store.typesSymbols[name]
}

func (store *javaBuiltinTypeStore) FindByName(name string) (lib.Symbol, bool) {
	if store.typesSymbols == nil {
		return nil, false
	}
	sym, ok := store.typesSymbols[name]
	return sym, ok
}

var builtinTypesStore = &javaBuiltinTypeStore{}

type ArraySymbol struct {
	ValueSymbol lib.Symbol
	Length      int
}

func (sym ArraySymbol) Name() string {
	return fmt.Sprintf("[%d]%s", sym.Length, sym.ValueSymbol.Name())
}

func (sym ArraySymbol) Kind() lib.SymbolKind {
	return lib.SymbolKindArray
}

func (sym ArraySymbol) Location() lib.Location {
	return sym.ValueSymbol.Location()
}

var BuiltinTypes = struct {
	NullSymbol    lib.Symbol
	BooleanSymbol lib.Symbol
	StringSymbol  lib.Symbol
	Integral      struct {
		ByteSymbol  lib.Symbol
		ShortSymbol lib.Symbol
		IntSymbol   lib.Symbol
		LongSymbol  lib.Symbol
		CharSymbol  lib.Symbol
	}
	FloatingPoint struct {
		FloatSymbol  lib.Symbol
		DoubleSymbol lib.Symbol
	}
	VoidSymbol lib.Symbol
}{
	NullSymbol:    builtinTypesStore.Builtin("null"),
	BooleanSymbol: builtinTypesStore.Builtin("boolean"),
	StringSymbol:  builtinTypesStore.Builtin("String"),
	Integral: struct {
		ByteSymbol  lib.Symbol
		ShortSymbol lib.Symbol
		IntSymbol   lib.Symbol
		LongSymbol  lib.Symbol
		CharSymbol  lib.Symbol
	}{
		ByteSymbol:  builtinTypesStore.Builtin("byte"),
		ShortSymbol: builtinTypesStore.Builtin("short"),
		IntSymbol:   builtinTypesStore.Builtin("int"),
		LongSymbol:  builtinTypesStore.Builtin("long"),
		CharSymbol:  builtinTypesStore.Builtin("char"),
	},
	FloatingPoint: struct {
		FloatSymbol  lib.Symbol
		DoubleSymbol lib.Symbol
	}{
		FloatSymbol:  builtinTypesStore.Builtin("float"),
		DoubleSymbol: builtinTypesStore.Builtin("double"),
	},
	VoidSymbol: builtinTypesStore.Builtin("void"),
}

func arrayIfy(typ lib.Symbol, len int) lib.Symbol {
	return ArraySymbol{
		ValueSymbol: typ,
		Length:      len,
	}
}
