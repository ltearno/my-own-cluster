#include <stdint.h>

#define WASM_EXPORT                   extern __attribute__((used)) __attribute__((visibility ("default")))
#define WASM_EXPORT_AS(NAME)          WASM_EXPORT __attribute__((export_name(NAME)))
#define WASM_IMPORT(MODULE,NAME)      __attribute__((import_module(MODULE))) __attribute__((import_name(NAME)))
#define WASM_CONSTRUCTOR              __attribute__((constructor))

WASM_IMPORT("rust-multiply", "rustMultiply")
uint32_t rustMultiply(uint32_t a, uint32_t b);

WASM_IMPORT("rust-multiply", "rustDivide")
uint32_t rustDivide(uint32_t a, uint32_t b);

WASM_EXPORT
uint32_t process(int a, int b) {
    return rustDivide(rustMultiply(a,a), rustMultiply(b,b));
}