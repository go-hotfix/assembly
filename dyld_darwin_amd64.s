//go:build darwin && amd64

#include "textflag.h"

TEXT ·dyldImageCount(SB),NOSPLIT,$0
    CALL _dyld_image_count(SB)
    MOVL AX, ret+0(FP)
    RET

TEXT ·dyldGetImageName(SB),NOSPLIT,$0
    MOVL index+0(FP), DI
    CALL _dyld_get_image_name(SB)
    MOVQ AX, ret+8(FP)
    RET

TEXT ·dyldGetImageHeader(SB),NOSPLIT,$0
    MOVL index+0(FP), DI
    CALL _dyld_get_image_header(SB)
    MOVQ AX, ret+8(FP)
    RET
