#include "my-own-cluster-guest-api.h"

WASM_IMPORT("rust-multiply", "rustMultiply")
uint32_t rustMultiply(uint32_t a, uint32_t b);

WASM_IMPORT("rust-multiply", "rustDivide")
uint32_t rustDivide(uint32_t a, uint32_t b);

WASM_IMPORT("api-demo-c", "get_size_of_passed_buffer")
uint32_t get_size_of_passed_buffer(int bufferId);

WASM_EXPORT
uint32_t process(int a, int b) {
    // this registers an exchange buffer and pass it through IFC to another program
    // the other program is running in another wasm instance and is automatically dynamically linked
    // the other program reads the exchange buffer and returns it size, which should be 5
    char buffer[] = "hello";
    int bufferId = register_buffer(buffer, 5);
    int v = get_size_of_passed_buffer(bufferId);
    
    return v + rustDivide(rustMultiply(a,a), rustMultiply(b,b));
}