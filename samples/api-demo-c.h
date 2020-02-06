#ifndef api_demo_api_h
#define api_demo_api_h

#include <stdint.h>

#define WASM_EXPORT                   extern __attribute__((used)) __attribute__((visibility ("default")))
#define WASM_EXPORT_AS(NAME)          WASM_EXPORT __attribute__((export_name(NAME)))
#define WASM_IMPORT(MODULE,NAME)      __attribute__((import_module(MODULE))) __attribute__((import_name(NAME)))
#define WASM_CONSTRUCTOR              __attribute__((constructor))

WASM_IMPORT("my-own-cluster", "test") uint32_t test();

WASM_IMPORT("my-own-cluster", "register_buffer") uint32_t register_buffer(void *buffer, int length);
WASM_IMPORT("my-own-cluster", "get_buffer_size") uint32_t get_buffer_size(int bufferId);
WASM_IMPORT("my-own-cluster", "get_buffer") uint32_t get_buffer(int bufferId, void *buffer, int length);
WASM_IMPORT("my-own-cluster", "free_buffer") uint32_t free_buffer(int bufferId);

#endif