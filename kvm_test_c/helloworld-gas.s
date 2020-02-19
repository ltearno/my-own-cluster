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
#   gcc -c helloworld-gas.s && ld helloworld-gas.o && objdump -D ./a.out && ./a.out
#
# ----------------------------------------------------------------------------------------

        .global _start

        .text
_start:
        # write(1, message, 13)
        mov     0x1000, %eax                # system call 1 is write
        movl    $42, (%eax)
