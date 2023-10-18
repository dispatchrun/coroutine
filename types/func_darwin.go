package types

import (
	"debug/macho"
	"fmt"
)

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
