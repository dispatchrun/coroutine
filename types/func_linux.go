package types

import (
	"debug/elf"
	"os"
)

func init() {
	f, err := elf.Open(os.Args[0])
	if err != nil {
		panic("cannot read elf binary: " + err.Error())
	}

	pclntab := f.Section(".gopclntab")
	pclntabData, err := readAll(pclntab, pclntab.Size)
	if err != nil {
		panic("cannot read pclntab: " + err.Error())
	}

	symtab := f.Section(".gosymtab")
	symtabData, err := readAll(symtab, symtab.Size)
	if err != nil {
		panic("cannot read symtab: " + err.Error())
	}

	initFunctionTables(pclntabData, symtabData)
}
