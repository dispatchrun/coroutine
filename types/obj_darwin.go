package types

import (
	"debug/macho"
	"os"
)

func init() {
	f, err := macho.Open(os.Args[0])
	if err != nil {
		panic("cannot read Mach-O binary: " + err.Error())
	}
	defer f.Close()

	initMachOFunctionTables(f)
	initMachOBuildID(f)
}

func initMachOFunctionTables(f *macho.File) {
	pclntab := f.Section("__gopclntab")
	pclntabData, err := readSection(pclntab, pclntab.Size)
	if err != nil {
		panic("cannot read pclntab: " + err.Error())
	}
	symtab := f.Section("__gosymtab")
	symtabData, err := readSection(symtab, symtab.Size)
	if err != nil {
		panic("cannot read symtab: " + err.Error())
	}
	initFunctionTables(pclntabData, symtabData)
}

func initMachOBuildID(f *macho.File) {
	text := f.Section("__text")

	// Read up to 32KB from the text section.
	// See https://github.com/golang/go/blob/3803c858/src/cmd/internal/buildid/note.go#L199
	buf, err := readSection(text, min(text.Size, 32*1024))
	if err != nil {
		panic("cannot read __text: " + err.Error())
	}
	if err := parseBuildID(buf); err != nil {
		panic(err)
	}
}
