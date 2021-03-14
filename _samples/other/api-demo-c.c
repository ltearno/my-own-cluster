#include "api-demo-guest-api.h"

#define NULL 0

WASM_EXPORT
uint32_t _start(int a, int b) {
    // call the host test function
    print_debug("hello", 5);

    /**
     * This demonstrates how to pass a memory buffer from guest to host and vice-versa.
     */

    // register a buffer in the host
    char buffer[] = "hello";
    int bufferId = create_exchange_buffer();
    write_exchange_buffer(bufferId, buffer, 5);

    // get buffer
    char dest[10];
    int bufferSize = read_exchange_buffer(bufferId, NULL, 0);
    if(bufferSize!=5) {
        return 100;
    }
    read_exchange_buffer(bufferId, dest, 5);

    // check buffer
    for(int i=0;i<5;i++){
        if(buffer[i] != dest[i]) {
            return i;
        }
    }

    // free buffer
    free_buffer(bufferId);

    return 700;
}

WASM_EXPORT
uint32_t get_size_of_passed_buffer(int bufferId) {
    return read_exchange_buffer(bufferId, NULL, 0);
}