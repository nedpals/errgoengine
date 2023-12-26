package python

import (
	"fmt"

	lib "github.com/nedpals/errgoengine"
)

type pythonBuiltinTypeStore struct {
	typesSymbols map[string]lib.Symbol
}

func (store *pythonBuiltinTypeStore) Builtin(name string) lib.Symbol {
	if store.typesSymbols == nil {
		store.typesSymbols = make(map[string]lib.Symbol)
	} else if sym, ok := store.FindByName(name); ok {
		return sym
	}
	store.typesSymbols[name] = lib.Builtin(name)
	return store.typesSymbols[name]
}

func (store *pythonBuiltinTypeStore) FindByName(name string) (lib.Symbol, bool) {
	if store.typesSymbols == nil {
		return nil, false
	}
	sym, ok := store.typesSymbols[name]
	return sym, ok
}

var builtinTypesStore = &pythonBuiltinTypeStore{}

// built-in types in python
var BuiltinTypes = struct {
	NoneSymbol    lib.Symbol
	AnySymbol     lib.Symbol
	VoidSymbol    lib.Symbol
	BooleanSymbol lib.Symbol
	StringSymbol  lib.Symbol
	IntSymbol     lib.Symbol
	FloatSymbol   lib.Symbol
}{
	NoneSymbol:    builtinTypesStore.Builtin("none"),
	AnySymbol:     builtinTypesStore.Builtin("any"),
	VoidSymbol:    builtinTypesStore.Builtin("void"),
	BooleanSymbol: builtinTypesStore.Builtin("bool"),
	StringSymbol:  builtinTypesStore.Builtin("str"),
	IntSymbol:     builtinTypesStore.Builtin("int"),
	FloatSymbol:   builtinTypesStore.Builtin("float"),
}

type TupleSymbol struct {
	Items []lib.Symbol
}

func (sym TupleSymbol) Name() string {
	return "tuple"
}

func (sym TupleSymbol) Kind() lib.SymbolKind {
	return lib.SymbolKindType
}

func (sym TupleSymbol) Location() lib.Location {
	return lib.Location{}
}

func (sym TupleSymbol) IsFixed() bool {
	return true
}

type BuiltinCollectionTypeKind int

func (k BuiltinCollectionTypeKind) String() string {
	switch k {
	case CollectionTypeList:
		return "list"
	case CollectionTypeDict:
		return "dict"
	case CollectionTypeTuple:
		return "tuple"
	case CollectionTypeSet:
		return "set"
	default:
		return "unknown"
	}
}

func (k BuiltinCollectionTypeKind) IsFixed() bool {
	return k == CollectionTypeTuple
}

var strToBuiltinCollectionTypeSyms = map[string]BuiltinCollectionTypeKind{
	CollectionTypeList.String():  CollectionTypeList,
	CollectionTypeDict.String():  CollectionTypeDict,
	CollectionTypeTuple.String(): CollectionTypeTuple,
	CollectionTypeSet.String():   CollectionTypeSet,
}

const (
	CollectionTypeList  BuiltinCollectionTypeKind = 0
	CollectionTypeDict  BuiltinCollectionTypeKind = iota
	CollectionTypeTuple BuiltinCollectionTypeKind = iota
	CollectionTypeSet   BuiltinCollectionTypeKind = iota
)

type CollectionTypeSymbol struct {
	TypeKind  BuiltinCollectionTypeKind
	KeyType   lib.Symbol
	ValueType []lib.Symbol
}

func (sym CollectionTypeSymbol) Name() string {
	return sym.TypeKind.String()
}

func (sym CollectionTypeSymbol) Kind() lib.SymbolKind {
	return lib.SymbolKindType
}

func (sym CollectionTypeSymbol) Location() lib.Location {
	return lib.Location{}
}

func (sym CollectionTypeSymbol) IsFixed() bool {
	return sym.TypeKind.IsFixed()
}

func collectionIfy(name string, syms ...lib.Symbol) (lib.Symbol, error) {
	kind, ok := strToBuiltinCollectionTypeSyms[name]
	if !ok {
		return nil, fmt.Errorf("not a valid collection type")
	}

	sym := CollectionTypeSymbol{
		TypeKind:  kind,
		KeyType:   BuiltinTypes.IntSymbol,
		ValueType: []lib.Symbol{BuiltinTypes.AnySymbol},
	}

	switch kind {
	case CollectionTypeList, CollectionTypeSet:
		if len(syms) == 1 {
			sym.ValueType[0] = syms[0]
		}
	case CollectionTypeDict:
		if len(syms) == 2 {
			sym.KeyType = syms[0]
			sym.ValueType[0] = syms[1]
		} else {
			sym.KeyType = BuiltinTypes.StringSymbol
		}
	case CollectionTypeTuple:
		sym.ValueType = syms
	}

	return sym, nil
}
