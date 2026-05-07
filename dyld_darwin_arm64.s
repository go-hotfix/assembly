//go:build darwin && arm64

#include "textflag.h"

TEXT ·dyldImageCount(SB),NOSPLIT,$0
    CALL _dyld_image_count(SB)
    MOVW R0, ret+0(FP)
    RET

TEXT ·dyldGetImageName(SB),NOSPLIT,$0
    MOVW index+0(FP), R0
    CALL _dyld_get_image_name(SB)
    MOVD R0, ret+8(FP)
    RET

TEXT ·dyldGetImageHeader(SB),NOSPLIT,$0
    MOVW index+0(FP), R0
    CALL _dyld_get_image_header(SB)
    MOVD R0, ret+8(FP)
    RET
