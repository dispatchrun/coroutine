package types

import (
	"debug/elf"
)

func initELFFunctionTables(f *elf.File) error {
	pclntab := f.Section(".gopclntab")
	pclntabData, err := readAll(pclntab, pclntab.Size)
	if err != nil {
		return fmt.Errorf("cannot read pclntab: %w", err)
	}
	symtab := f.Section(".gosymtab")
	symtabData, err := readAll(symtab, symtab.Size)
	if err != nil {
		return fmt.Errorf("cannot read symtab: %w", err)
	}
	initFunctionTables(pclntabData, symtabData)
	return nil
}
