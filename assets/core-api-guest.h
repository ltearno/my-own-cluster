
#ifndef core_api_h
#define core_api_h

#include <stdint.h>

#define WASM_EXPORT                   extern __attribute__((used)) __attribute__((visibility ("default")))
#define WASM_EXPORT_AS(NAME)          WASM_EXPORT __attribute__((export_name(NAME)))
#define WASM_IMPORT(MODULE,NAME)      __attribute__((import_module(MODULE))) __attribute__((import_name(NAME)))
#define WASM_CONSTRUCTOR              __attribute__((constructor))

WASM_IMPORT("core", "get_input_buffer_id") uint32_t get_input_buffer_id();
WASM_IMPORT("core", "get_output_buffer_id") uint32_t get_output_buffer_id();
WASM_IMPORT("core", "create_exchange_buffer") uint32_t create_exchange_buffer();
WASM_IMPORT("core", "write_exchange_buffer") uint32_t write_exchange_buffer(int buffer_id, const void *content_bytes, int content_length);
WASM_IMPORT("core", "write_exchange_buffer_header") uint32_t write_exchange_buffer_header(int buffer_id, const char *name_string, int name_length, const char *value_string, int value_length);
WASM_IMPORT("core", "write_exchange_buffer_status_code") uint32_t write_exchange_buffer_status_code(int buffer_id, int status_code);
// returns the readden size if result_bytes was not NULL and the exchange buffer size otherwise
WASM_IMPORT("core", "read_exchange_buffer") uint32_t read_exchange_buffer(int buffer_id, void *result_bytes, int result_length);
// returns the buffer headers in JSON format
WASM_IMPORT("core", "read_exchange_buffer_headers") uint32_t read_exchange_buffer_headers(int buffer_id);
WASM_IMPORT("core", "base64_decode") uint32_t base64_decode(const char *encoded_string, int encoded_length);
WASM_IMPORT("core", "base64_encode") uint32_t base64_encode(const void *input_bytes, int input_length);
WASM_IMPORT("core", "register_blob_with_name") uint32_t register_blob_with_name(const char *name_string, int name_length, const char *content_type_string, int content_type_length, const void *content_bytes, int content_length);
WASM_IMPORT("core", "register_blob") uint32_t register_blob(const char *content_type_string, int content_type_length, const void *content_bytes, int content_length);
WASM_IMPORT("core", "get_blob_tech_id_from_name") uint32_t get_blob_tech_id_from_name(const char *name_string, int name_length);
WASM_IMPORT("core", "get_blob_bytes_as_string") uint32_t get_blob_bytes_as_string(const char *name_string, int name_length);
WASM_IMPORT("core", "plug_function") uint32_t plug_function(const char *method_string, int method_length, const char *path_string, int path_length, const char *name_string, int name_length, const char *start_function_string, int start_function_length, const char *data_string, int data_length);
WASM_IMPORT("core", "plug_file") uint32_t plug_file(const char *method_string, int method_length, const char *path_string, int path_length, const char *name_string, int name_length);
WASM_IMPORT("core", "unplug_path") uint32_t unplug_path(const char *method_string, int method_length, const char *path_string, int path_length);
WASM_IMPORT("core", "get_status") uint32_t get_status();
WASM_IMPORT("core", "persistence_set") uint32_t persistence_set(const void *key_bytes, int key_length, const void *value_bytes, int value_length);
WASM_IMPORT("core", "get_url") uint32_t get_url(const char *url_string, int url_length);
WASM_IMPORT("core", "persistence_get") uint32_t persistence_get(const void *key_bytes, int key_length);
WASM_IMPORT("core", "persistence_get_subset") uint32_t persistence_get_subset(const char *prefix_string, int prefix_length);
WASM_IMPORT("core", "print_debug") uint32_t print_debug(const char *text_string, int text_length);
WASM_IMPORT("core", "get_time") uint32_t get_time(const void *dest_bytes, int dest_length);
WASM_IMPORT("core", "free_buffer") uint32_t free_buffer(int bufferId);
WASM_IMPORT("core", "call_function") uint32_t call_function(const char *name_string, int name_length, const char *start_function_string, int start_function_length, const void *arguments_int_array, int arguments_length, const char *mode_string, int mode_length, int input_exchange_buffer_id, int output_exchange_buffer_id, const char *posix_file_name_string, int posix_file_name_length, const void *posix_arguments_string_array, int posix_arguments_length);
WASM_IMPORT("core", "export_database") uint32_t export_database();
WASM_IMPORT("core", "beta_web_proxy") uint32_t beta_web_proxy(const char *proxy_spec_json_string, int proxy_spec_json_length);

#endif
    