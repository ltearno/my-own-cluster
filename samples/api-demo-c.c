#include "api-demo-c.h"

WASM_EXPORT
uint32_t _start(int a, int b) {
    test();
    return a+b;
}