        .global _start

        .text
_start:
        push $1
        push $2
        push $3
        push $4

        # point to the address 0x100
        mov $0x100, %rbx

        # read memory to rax
        movq (%rbx), %rax

        # increment rax
        inc %rax

        # and write in memory at the same place
        movq %rax, (%rbx)

        hlt
