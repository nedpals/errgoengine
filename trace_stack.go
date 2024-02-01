package errgoengine

import "strings"

type TraceStack []StackTraceEntry

func (st *TraceStack) Add(symbolName string, loc Location) {
	*st = append(*st, StackTraceEntry{
		Location:   loc,
		SymbolName: symbolName,
	})
}

func (st TraceStack) Top() StackTraceEntry {
	if len(st) == 0 {
		return StackTraceEntry{}
	}
	return st[len(st)-1]
}

func (st TraceStack) NearestTo(path string) StackTraceEntry {
	for i := len(st) - 1; i >= 0; i-- {
		entry := st[i]
		if !strings.HasPrefix(entry.DocumentPath, path) {
			continue
		}
		return entry
	}
	return st.Top()
}

type StackTraceEntry struct {
	Location
	SymbolName string
}
