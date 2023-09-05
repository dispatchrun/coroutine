#include "go_asm.h"
#include "go_tls.h"
#include "testflag.h"

// GOARCH=arm64 exposes the get_tls and g macros to access thread local storage
// and the g pointer in it.
//
// See: https://go.dev/doc/asm (64-bit Intel 386)
TEXT ·getg(SB), NOSPLIT, $0-8
    get_tls(CX)
    MOVQ g(CX), AX
    MOVQ AX, ret+0(FP)
    RET
