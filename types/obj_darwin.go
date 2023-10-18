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

	if err := initMachOFunctionTables(f); err != nil {
		panic(err)
	}
}
