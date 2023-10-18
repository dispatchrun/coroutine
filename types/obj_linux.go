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
