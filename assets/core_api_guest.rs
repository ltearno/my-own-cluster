
use std::collections::HashMap;
use std::io::{Cursor, Read};
use byteorder::{LittleEndian, ReadBytesExt};

/**
 * 'core' guest API bindings for rust
 * 
 * It can be called from rust code compiled into wasm and executed by my-own-cluster
 */

// import the 'core' module
pub mod raw {
    #[link(wasm_import_module = "core")]
    extern {
        pub fn get_input_buffer_id() -> u32;
        pub fn get_output_buffer_id() -> u32;
        pub fn create_exchange_buffer() -> u32;
        pub fn write_exchange_buffer(buffer_id:u32, content_bytes: *const u8, content_length: u32) -> u32;
        pub fn write_exchange_buffer_header(buffer_id:u32, name_string: *const u8, name_length: u32, value_string: *const u8, value_length: u32) -> u32;
        pub fn get_exchange_buffer_size(buffer_id:u32) -> u32;
        // returns the written size if result_bytes was not NULL and the exchange buffer size otherwise
        pub fn read_exchange_buffer(buffer_id:u32, result_bytes: *mut u8, result_length: u32) -> u32;
        // returns the buffer headers in JSON format
        pub fn read_exchange_buffer_headers(buffer_id:u32) -> u32;
        pub fn get_buffer_headers(buffer_id:u32) -> u32;
        pub fn base64_decode(encoded_string: *const u8, encoded_length: u32) -> u32;
        pub fn base64_encode(input_bytes: *const u8, input_length: u32) -> u32;
        pub fn register_blob_with_name(name_string: *const u8, name_length: u32, content_type_string: *const u8, content_type_length: u32, content_bytes: *const u8, content_length: u32) -> u32;
        pub fn register_blob(content_type_string: *const u8, content_type_length: u32, content_bytes: *const u8, content_length: u32) -> u32;
        pub fn get_blob_tech_id_from_name(name_string: *const u8, name_length: u32) -> u32;
        pub fn get_blob_bytes_as_string(name_string: *const u8, name_length: u32) -> u32;
        pub fn plug_function(method_string: *const u8, method_length: u32, path_string: *const u8, path_length: u32, name_string: *const u8, name_length: u32, start_function_string: *const u8, start_function_length: u32) -> u32;
        pub fn plug_file(method_string: *const u8, method_length: u32, path_string: *const u8, path_length: u32, name_string: *const u8, name_length: u32) -> u32;
        pub fn get_status() -> u32;
        pub fn persistence_set(key_bytes: *const u8, key_length: u32, value_bytes: *const u8, value_length: u32) -> u32;
        pub fn get_url(url_string: *const u8, url_length: u32) -> u32;
        pub fn persistence_get(key_bytes: *const u8, key_length: u32) -> u32;
        pub fn persistence_get_subset(prefix_string: *const u8, prefix_length: u32) -> u32;
        pub fn print_debug(text_string: *const u8, text_length: u32) -> u32;
        pub fn get_time(dest_bytes: *const u8, dest_length: u32) -> u32;
        pub fn free_buffer(bufferId:u32) -> u32;
        pub fn call_function(name_string: *const u8, name_length: u32, start_function_string: *const u8, start_function_length: u32, arguments_int_array: *const u32, arguments_length: u32, mode_string: *const u8, mode_length: u32, input_exchange_buffer_id:u32, output_exchange_buffer_id:u32, posix_file_name_string: *const u8, posix_file_name_length: u32, posix_arguments_string_array: *const u8, posix_arguments_length: u32) -> u32;

    }
}
    
    