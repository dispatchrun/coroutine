package closures

import (
	"debug/macho"
	"os"
)

func init() {
	f, err := macho.Open(os.Args[0])
	if err != nil {
		panic("cannot read Mach-O binary: " + err.Error())
	}

	pclntab := f.Section("__gopclntab")
	pclntabData, err := readAll(pclntab, pclntab.Size)
	if err != nil {
		panic("cannot read pclntab: " + err.Error())
	}

	symtab := f.Section("__gosymtab")
	symtabData, err := readAll(symtab, symtab.Size)
	if err != nil {
		panic("cannot read symtab: " + err.Error())
	}

	initFunctionTables(pclntabData, symtabData)
}
