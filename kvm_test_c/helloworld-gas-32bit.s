# ----------------------------------------------------------------------------------------
# Writes "Hello, World" to the console using only system calls. Runs on 64-bit Linux only.
# To assemble and run:
#
#     gcc -c hello.s && ld hello.o && ./a.out
#
# or
#
#     gcc -nostdlib hello.s && ./a.out
#
#   gcc -m32 -c helloworld-gas-32bit.s && objdump -D helloworld-gas-32bit.o
#
# ----------------------------------------------------------------------------------------

        .global _start

        .text
        .code32
_start:
        # write(1, message, 13)
        mov    $0x10, %eax                # system call 1 is write
        movw    $42, (%eax)
