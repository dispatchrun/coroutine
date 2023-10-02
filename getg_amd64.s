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

TEXT ·getg(SB), NOSPLIT, $0-8
    MOVQ TLS, CX
    MOVQ 0(CX)(TLS*1), BX // g
    MOVQ BX, ret+0(FP)
    RET
