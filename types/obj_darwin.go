package types

import (
	"debug/macho"
	"fmt"
	"os"
)

func init() {
	f, err := macho.Open(os.Args[0])
	if err != nil {
		panic("cannot read Mach-O binary: " + err.Error())
	}
	defer f.Close()

	if err := initMachOFunctionTables(f); err != nil {
		panic(err)
	}
}

func initMachOFunctionTables(f *macho.File) error {
	pclntab := f.Section("__gopclntab")
	pclntabData, err := readAll(pclntab, pclntab.Size)
	if err != nil {
		return fmt.Errorf("cannot read pclntab: %w", err)
	}
	symtab := f.Section("__gosymtab")
	symtabData, err := readAll(symtab, symtab.Size)
	if err != nil {
		return fmt.Errorf("cannot read symtab: %w", err)
	}
	initFunctionTables(pclntabData, symtabData)
	return nil
}
