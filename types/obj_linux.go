package types

import (
	"bytes"
	"debug/elf"
	"os"
)

func init() {
	f, err := elf.Open(os.Args[0])
	if err != nil {
		panic("cannot read elf binary: " + err.Error())
	}
	defer f.Close()

	initELFFunctionTables(f)
	initELFBuildID(f)
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

func initELFBuildID(f *elf.File) {
	noteSection := f.Section(".note.go.buildid")
	note, err := readSection(noteSection, noteSection.Size)
	if err != nil {
		panic("cannot read build ID: " + err.Error())
	}

	// See https://github.com/golang/go/blob/3803c858/src/cmd/internal/buildid/note.go#L135C3-L135C3
	nameSize := f.ByteOrder.Uint32(note)
	valSize := f.ByteOrder.Uint32(note[4:])
	tag := f.ByteOrder.Uint32(note[8:])
	nname := note[12:16]
	if nameSize == 4 && 16+valSize <= uint32(len(note)) && tag == buildIDTag && bytes.Equal(nname, buildIDNote) {
		buildid = string(note[16 : 16+valSize])
	} else {
		panic("build ID not found")
	}
}

var (
	buildIDNote = []byte("Go\x00\x00")
	buildIDTag  = uint32(4)
)
