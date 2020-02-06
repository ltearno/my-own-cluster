#include "api-demo-c.h"

WASM_EXPORT
uint32_t _start(int a, int b) {
    // call the host test function
    test();

    /**
     * This demonstrates how to pass a memory buffer from guest to host and vice-versa.
     */

    // register a buffer in the host
    char buffer[] = "hello";
    int bufferId = register_buffer(buffer, 5);

    // get buffer
    char dest[10];
    int bufferSize = get_buffer_size(bufferId);
    if(bufferSize!=5) {
        return 100;
    }
    get_buffer(bufferId, dest, 5);

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