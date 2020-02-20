        .global _start

        .text
_start:
        xor %rax, %rax
        mov $7, %rax
        mov $0xee3578622345678, %rax
        mov $7, %rbx
        movq (%rbx), %rax
        inc %rbx
        movq $5, (%rbx)
        hlt
