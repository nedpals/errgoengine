package java

import (
	"fmt"

	lib "github.com/nedpals/errgoengine"
)

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
	NullSymbol:    lib.Builtin("null"),
	BooleanSymbol: lib.Builtin("boolean"),
	StringSymbol:  lib.Builtin("string"),
	Integral: struct {
		ByteSymbol  lib.Symbol
		ShortSymbol lib.Symbol
		IntSymbol   lib.Symbol
		LongSymbol  lib.Symbol
		CharSymbol  lib.Symbol
	}{
		ByteSymbol:  lib.Builtin("byte"),
		ShortSymbol: lib.Builtin("short"),
		IntSymbol:   lib.Builtin("int"),
		LongSymbol:  lib.Builtin("long"),
		CharSymbol:  lib.Builtin("char"),
	},
	FloatingPoint: struct {
		FloatSymbol  lib.Symbol
		DoubleSymbol lib.Symbol
	}{
		FloatSymbol:  lib.Builtin("float"),
		DoubleSymbol: lib.Builtin("double"),
	},
	VoidSymbol: lib.Builtin("void"),
}

func arrayIfy(typ lib.Symbol, len int) lib.Symbol {
	return ArraySymbol{
		ValueSymbol: typ,
		Length:      len,
	}
}
