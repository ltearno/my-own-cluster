# This is a demonstration of 16 bit code reading and writing to memory

.global _start
    .code16

_start:
    # store 0x07 in the BX register
    mov $7, %bx

    # reads the value in memory at address given by BX (which is normally 0x07) into AX
    movb (%bx), %ax

    # write 0x05 in memory at address given by BX (which is normally 0x07)
    movb $5, (%bx)

    # halt the processor
    hlt
