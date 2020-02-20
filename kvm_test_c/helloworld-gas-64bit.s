        .global _start

        .text
_start:
        # mov     0x9876543212345678, %rax
        xor %rax, %rax
        mov $7, %rax
        hlt
