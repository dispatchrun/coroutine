#include "go_asm.h"
#include "textflag.h"

// GOARCH=386 exposes the get_tls and g macros to access thread local storage
// and the g pointer in it. The routine here inlines the following macros
// defined in $GOROOT/src/runtime/go_tls.h:
//
//  #define get_tls(r) MOVL TLS, r
//  #define g(r)       0(r)(TLS*1)
//
// See: https://go.dev/doc/asm (32-bit Intel 386)
TEXT ·getg(SB), NOSPLIT, $0-4
    MOVL TLS, CX
    MOVL 0(CX)(TLS*1), AX
    MOVL AX, ret+0(FP)
    RET
