package closures

import (
	"debug/gosym"
	"debug/macho"
	"os"
	"runtime"
)

func init() {
	f, err := macho.Open(os.Args[0])
	if err != nil {
		panic("cannot read Mach-O binary: " + err.Error())
	}

	pclntab := f.Section("__gopclntab")
	pclntabData := make([]byte, pclntab.Size)
	if n, err := pclntab.ReadAt(pclntabData, 0); err != nil && n < len(pclntabData) {
		panic("cannot read pclntab: " + err.Error())
	}

	symtab := f.Section("__gosymtab")
	symtabData := make([]byte, symtab.Size)
	if n, err := symtab.ReadAt(symtabData, 0); err != nil && n < len(symtabData) {
		panic("cannot read symtab: " + err.Error())
	}

	table, err := gosym.NewTable(symtabData, gosym.NewLineTable(pclntabData, 0))
	if err != nil {
		panic("cannot read symtab: " + err.Error())
	}

	sentinelName, sentinelAddr := sentinel()

	tableFunc := table.LookupFunc(sentinelName)
	offset := uint64(sentinelAddr) - tableFunc.Entry

	functions := make([]Func, len(table.Funcs))
	for i, fn := range table.Funcs {
		functions[i] = Func{
			Addr: uintptr(fn.Entry + offset),
			Name: fn.Name,
		}
	}

	initFunctionTables(functions)
}

//go:noinline
func sentinel() (name string, addr uintptr) {
	pc := [1]uintptr{}
	runtime.Callers(0, pc[:])

	fn := runtime.FuncForPC(pc[0])
	return fn.Name(), fn.Entry()
}
