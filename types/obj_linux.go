package types

func init() {
	f, err := elf.Open(os.Args[0])
	if err != nil {
		panic("cannot read elf binary: " + err.Error())
	}
	defer f.Close()

	if err := initELFFunctionTables(f); err != nil {
		panic(err)
	}
}

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
