#include "go_asm.h"
#include "textflag.h"

// GOARCH=amd64 exposes the get_tls and g macros to access thread local storage
// and the g pointer in it. The routine here inlines the following macros
// defined in $GOROOT/src/runtime/go_tls.h:
//
//  #define get_tls(r) MOVQ TLS, r
//  #define g(r)       0(r)(TLS*1)
//
// See: https://go.dev/doc/asm (64-bit Intel 386)

TEXT ·with(SB), NOSPLIT, $0-24
    MOVQ TLS, CX
    MOVQ 0(CX)(TLS*1), BX // g
    MOVQ 8(BX), AX        // g.stack.hi

    // The v argument is pushed on the stack by the caller, we use its offset
    // to the goroutine's stack pointer as key to relocate the value.
    LEAQ v_type+0(FP), CX
    SUBQ AX, CX // offset

    // Write the offset of v on the stack, this is used to relocate v in calls
    // to load.
    //
    // On amd64, the g struct is 408 bytes, but allocated on the heap it uses a
    // class size of 416 bytes, which means that we have 8 bytes unused at the
    // end of the struct where we can store the offset.
    MOVQ CX, 408(BX)

    MOVQ f+16(FP), AX
    MOVQ AX, DX // calling convention for closures
    CALL (AX)
    RET

TEXT ·load(SB), NOSPLIT, $0-16
    MOVQ TLS, CX
    MOVQ 0(CX)(TLS*1), BX // g
    MOVQ 8(BX), AX        // g.stack.hi
    MOVQ 408(BX), CX

    MOVQ 0(AX)(CX*1), R8
    MOVQ 8(AX)(CX*1), R9

    MOVQ R8, ret_type+0(FP)
    MOVQ R9, ret_data+8(FP)
    RET
