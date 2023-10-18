package types

import (
	"debug/elf"
	"os"
)

func init() {
	obj, err := elf.Open(os.Args[0])
	if err != nil {
		panic("cannot read elf binary: " + err.Error())
	}
	defer obj.Close()

	initELFFunctionTables(obj)
}

func initELFFunctionTables(f *elf.File) {
	pclntab := f.Section(".gopclntab")
	pclntabData, err := readSection(pclntab, pclntab.Size)
	if err != nil {
		panic("cannot read pclntab: " + err.Error())
	}
	symtab := f.Section(".gosymtab")
	symtabData, err := readSection(symtab, symtab.Size)
	if err != nil {
		panic("cannot read symtab: " + err.Error())
	}
	initFunctionTables(pclntabData, symtabData)
}
