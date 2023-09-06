#include "go_asm.h"
#include "textflag.h"

// On most platforms, Go dedicates a register to storing the current g; if we
// misused this approach we should get an error saying the symbol g is not
// defined.
//
// https://github.com/golang/go/blob/master/src/cmd/asm/internal/arch/arch.go
TEXT ·getg(SB), NOSPLIT, $0-8
    MOVD g, ret+0(FP)
    RET
