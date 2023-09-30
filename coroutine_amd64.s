#include "go_asm.h"
#include "textflag.h"

// func with(k *uintptr, v any, f func())
TEXT ·with(SB), NOSPLIT, $0-32
    MOVQ TLS, CX
    MOVQ 0(CX)(TLS*1), BX // g
    MOVQ 8(BX), AX        // g.stack.hi

    // The v argument is pushed on the stack by the caller, we use its offset
    // to the goroutine's stack pointer as key to reloacte the value.
    LEAQ v_type+8(FP), CX
    SUBQ AX, CX // key

    // Write the key associated with the value; this may race when called from
    // multiple goroutines but we don't care because we always write the same
    // value as long as a single key is associated with call site.
    MOVQ k+0(FP), BX
    MOVQ CX, (BX)

    MOVQ f+24(FP), BX
    MOVQ BX, DX // calling convention for closures
    CALL (BX)
    RET

// func load(k uintptr) any
TEXT ·load(SB), NOSPLIT, $0-24
    MOVQ TLS, CX
    MOVQ 0(CX)(TLS*1), BX // g
    MOVQ 8(BX), AX        // g.stack.hi

    MOVQ k+0(FP), CX
    MOVQ 0(AX)(CX*1), BX
    MOVQ 8(AX)(CX*1), CX

    MOVQ BX, ret_type+8(FP)
    MOVQ CX, ret_data+16(FP)
    RET
