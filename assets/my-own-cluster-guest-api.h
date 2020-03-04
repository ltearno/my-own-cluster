#ifndef api_demo_api_h
#define api_demo_api_h

#include <stdint.h>

#define WASM_EXPORT                   extern __attribute__((used)) __attribute__((visibility ("default")))
#define WASM_EXPORT_AS(NAME)          WASM_EXPORT __attribute__((export_name(NAME)))
#define WASM_IMPORT(MODULE,NAME)      __attribute__((import_module(MODULE))) __attribute__((import_name(NAME)))
#define WASM_CONSTRUCTOR              __attribute__((constructor))

WASM_IMPORT("my-own-cluster", "print_debug") uint32_t print_debug(void *buffer, int length);
WASM_IMPORT("my-own-cluster", "create_exchange_buffer") uint32_t create_exchange_buffer();
WASM_IMPORT("my-own-cluster", "write_exchange_buffer") uint32_t write_exchange_buffer(int bufferId, void *buffer, int length);
WASM_IMPORT("my-own-cluster", "read_exchange_buffer") uint32_t read_exchange_buffer(int bufferId, void *buffer, int length);
WASM_IMPORT("my-own-cluster", "free_buffer") uint32_t free_buffer(int bufferId);
WASM_IMPORT("my-own-cluster", "get_input_buffer_id") uint32_t get_input_buffer_id();
WASM_IMPORT("my-own-cluster", "get_output_buffer_id") uint32_t get_output_buffer_id();
WASM_IMPORT("my-own-cluster", "get_url") uint32_t get_url(char *url, int length);

#endif