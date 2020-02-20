// gcc -nostdlib -m64 test-64bit.c

void _start() {
    unsigned char* a = (unsigned char*) 0x100;
    *a = 0x42;
}