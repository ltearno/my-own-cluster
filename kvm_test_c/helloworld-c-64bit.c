/**
 * This program executes in ring 0
 * 
 * There is no standard library here, only bare metal.
 * 
 * About bare metal, there is nothing here, just a cpu
 * and some memory (no pci bus, bios or nothing).
 * The CPU executes in 64 bit mode.
 * It is not corresponding to some real machine but 
 * a very bare virtualized environment created with kvm.
*/

// a function that we can call because the stack pointer is
// correctly set by the vm host
unsigned char add(unsigned char a, unsigned char b) {
    return a + b;
}

// this is what will be called by the vm host
void _start() {
    // declare a pointer valued to an arbitray address,
    // depending on how memory is paginated, the address may
    // or not be accessible.
    // here we point to last byte of the first 1GB
    unsigned char* a = (unsigned char*) 0x3FFFFFFF;

    // read at the address
    unsigned char r = add(*a, *a);

    // write in that address
    *a = r;

    // stop the machine ;)
    asm ("hlt"   :)     ;
}