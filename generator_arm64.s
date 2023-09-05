#include "go_asm.h"
#include "textflag.h"

// GOARCH=arm64 renames R28 to g and uses this register to track the pointer to
// the current goroutine object.
//
// // Avoid unintentionally clobbering g using R28.
// delete(register, "R28")
// register["g"] = arm64.REG_R28
//
// See: src/cmd/asm/internal/arch/arch.go
TEXT ·getg(SB), NOSPLIT, $0-8
    MOVD g, ret+0(FP)
    RET
