
#ifndef gpu_api_h
#define gpu_api_h

#include <stdint.h>

#define WASM_EXPORT                   extern __attribute__((used)) __attribute__((visibility ("default")))
#define WASM_EXPORT_AS(NAME)          WASM_EXPORT __attribute__((export_name(NAME)))
#define WASM_IMPORT(MODULE,NAME)      __attribute__((import_module(MODULE))) __attribute__((import_name(NAME)))
#define WASM_CONSTRUCTOR              __attribute__((constructor))

WASM_IMPORT("gpu", "compute_shader") uint32_t compute_shader(const char *specification_string, int specification_length);
WASM_IMPORT("gpu", "create_image_from_rgba_float_pixels") uint32_t create_image_from_rgba_float_pixels(int width, int height, int pixels_exchange_buffer_id, int png_exchange_buffer_id);
WASM_IMPORT("gpu", "create_image_from_r_float_pixels") uint32_t create_image_from_r_float_pixels(int width, int height, int pixels_exchange_buffer_id, int png_exchange_buffer_id);

#endif
    