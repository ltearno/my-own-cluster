
use std::collections::HashMap;
use std::io::{Cursor, Read};
use byteorder::{LittleEndian, ReadBytesExt};

/**
 * 'core' guest API bindings for rust
 * 
 * it can be called from rust code compiled into wasm and executed by my-own-cluster
 * 
 * you can use it by adding that in your rust source files at the beginning :
 * 
 *  mod core_api_guest;
 *  use core_api_guest::*;
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
        pub fn base64_decode(encoded_string: *const u8, encoded_length: u32) -> u32;
        pub fn base64_encode(input_bytes: *const u8, input_length: u32) -> u32;
        pub fn register_blob_with_name(name_string: *const u8, name_length: u32, content_type_string: *const u8, content_type_length: u32, content_bytes: *const u8, content_length: u32) -> u32;
        pub fn register_blob(content_type_string: *const u8, content_type_length: u32, content_bytes: *const u8, content_length: u32) -> u32;
        pub fn get_blob_tech_id_from_name(name_string: *const u8, name_length: u32) -> u32;
        pub fn get_blob_bytes_as_string(name_string: *const u8, name_length: u32) -> u32;
        pub fn plug_function(method_string: *const u8, method_length: u32, path_string: *const u8, path_length: u32, name_string: *const u8, name_length: u32, start_function_string: *const u8, start_function_length: u32) -> u32;
        pub fn plug_file(method_string: *const u8, method_length: u32, path_string: *const u8, path_length: u32, name_string: *const u8, name_length: u32) -> u32;
        pub fn unplug_path(method_string: *const u8, method_length: u32, path_string: *const u8, path_length: u32) -> u32;
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

pub fn get_input_buffer_id() -> u32 {
    unsafe { raw::get_input_buffer_id() }
}

pub fn get_output_buffer_id() -> u32 {
    unsafe { raw::get_output_buffer_id() }
}

pub fn create_exchange_buffer() -> u32 {
    unsafe { raw::create_exchange_buffer() }
}

pub fn write_exchange_buffer(buffer_id:u32, content: &[u8]) -> u32 {
    unsafe { raw::write_exchange_buffer(buffer_id, content.as_ptr(), content.len() as u32) }
}

pub fn write_exchange_buffer_header(buffer_id:u32, name: &str, value: &str) -> u32 {
    unsafe { raw::write_exchange_buffer_header(buffer_id, name.as_bytes().as_ptr(), name.as_bytes().len() as u32, value.as_bytes().as_ptr(), value.as_bytes().len() as u32) }
}

pub fn get_exchange_buffer_size(buffer_id:u32) -> u32 {
    unsafe { raw::get_exchange_buffer_size(buffer_id) }
}

pub fn read_exchange_buffer(buffer_id:u32) -> Result<Vec<u8>, u32> {
    let result_size = unsafe { raw::read_exchange_buffer(buffer_id, std::ptr::null_mut(), 0) };
    let mut result = Vec::with_capacity(result_size as usize);
    result.resize(result_size as usize, 0);
    unsafe { raw::read_exchange_buffer(buffer_id, result.as_mut_ptr(), result_size) };
    Ok(result)
}

pub fn read_exchange_buffer_headers(buffer_id:u32) -> Result<HashMap<String,String>, u32> {
    let result_buffer_id = unsafe { raw::read_exchange_buffer_headers(buffer_id) };
    if result_buffer_id == 0xffff {
        Err(5)
    }
    else {
        let result_buffer = read_exchange_buffer(result_buffer_id);
        match result_buffer {
            Ok(result_buffer) => {
                let mut result = HashMap::new();
                let buffer = result_buffer.as_slice();
                free_buffer(result_buffer_id);

                let mut rdr = Cursor::new(buffer);
                let mut current_key : String = "".to_string();

                let mut nb_buffers = rdr.read_u32::<LittleEndian>().unwrap();
                while nb_buffers > 0 {
                    let buffer_size = rdr.read_u32::<LittleEndian>().unwrap();
                    
                    let mut buffer = Vec::with_capacity(buffer_size as usize);
                    unsafe { buffer.set_len(buffer_size as usize); }
                    rdr.read(&mut buffer);

                    let s = String::from_utf8(buffer.to_vec()).unwrap();
                    
                    if nb_buffers % 2 == 0 {
                        current_key = s;
                    } else {
                        result.insert(current_key.clone(), s);
                    }

                    nb_buffers = nb_buffers - 1;        
                }

                Ok(result)
            },
            Err(err) => {
                Err(6)
            },
        }
    }
}

pub fn base64_decode(encoded: &str) -> Result<Vec<u8>, u32> {
    let result_buffer_id = unsafe { raw::base64_decode(encoded.as_bytes().as_ptr(), encoded.as_bytes().len() as u32) };
    if result_buffer_id == 0xffff {
        Err(1)
    }
    else {
        let result = read_exchange_buffer(result_buffer_id);
        match result {
            Ok(result) => {
                //free_buffer(result_buffer_id);
                Ok(result)
            },
            Err(err) => {
                Err(2)
            },
        }
    }
}

pub fn base64_encode(input: &[u8]) -> Result<String, u32> {
    let result_buffer_id = unsafe { raw::base64_encode(input.as_ptr(), input.len() as u32) };
    if result_buffer_id == 0xffff {
        Err(3)
    }
    else {
        let result_buffer = read_exchange_buffer(result_buffer_id);
        match result_buffer {
            Ok(result_buffer) => {
                let result = String::from_utf8(result_buffer).unwrap();
                //free_buffer(result_buffer_id);
                Ok(result)
            },
            Err(err) => {
                Err(4)
            },
        }
    }
}

pub fn register_blob_with_name(name: &str, content_type: &str, content: &[u8]) -> Result<String, u32> {
    let result_buffer_id = unsafe { raw::register_blob_with_name(name.as_bytes().as_ptr(), name.as_bytes().len() as u32, content_type.as_bytes().as_ptr(), content_type.as_bytes().len() as u32, content.as_ptr(), content.len() as u32) };
    if result_buffer_id == 0xffff {
        Err(3)
    }
    else {
        let result_buffer = read_exchange_buffer(result_buffer_id);
        match result_buffer {
            Ok(result_buffer) => {
                let result = String::from_utf8(result_buffer).unwrap();
                //free_buffer(result_buffer_id);
                Ok(result)
            },
            Err(err) => {
                Err(4)
            },
        }
    }
}

pub fn register_blob(content_type: &str, content: &[u8]) -> Result<String, u32> {
    let result_buffer_id = unsafe { raw::register_blob(content_type.as_bytes().as_ptr(), content_type.as_bytes().len() as u32, content.as_ptr(), content.len() as u32) };
    if result_buffer_id == 0xffff {
        Err(3)
    }
    else {
        let result_buffer = read_exchange_buffer(result_buffer_id);
        match result_buffer {
            Ok(result_buffer) => {
                let result = String::from_utf8(result_buffer).unwrap();
                //free_buffer(result_buffer_id);
                Ok(result)
            },
            Err(err) => {
                Err(4)
            },
        }
    }
}

pub fn get_blob_tech_id_from_name(name: &str) -> Result<String, u32> {
    let result_buffer_id = unsafe { raw::get_blob_tech_id_from_name(name.as_bytes().as_ptr(), name.as_bytes().len() as u32) };
    if result_buffer_id == 0xffff {
        Err(3)
    }
    else {
        let result_buffer = read_exchange_buffer(result_buffer_id);
        match result_buffer {
            Ok(result_buffer) => {
                let result = String::from_utf8(result_buffer).unwrap();
                //free_buffer(result_buffer_id);
                Ok(result)
            },
            Err(err) => {
                Err(4)
            },
        }
    }
}

pub fn get_blob_bytes_as_string(name: &str) -> Result<String, u32> {
    let result_buffer_id = unsafe { raw::get_blob_bytes_as_string(name.as_bytes().as_ptr(), name.as_bytes().len() as u32) };
    if result_buffer_id == 0xffff {
        Err(3)
    }
    else {
        let result_buffer = read_exchange_buffer(result_buffer_id);
        match result_buffer {
            Ok(result_buffer) => {
                let result = String::from_utf8(result_buffer).unwrap();
                //free_buffer(result_buffer_id);
                Ok(result)
            },
            Err(err) => {
                Err(4)
            },
        }
    }
}

pub fn plug_function(method: &str, path: &str, name: &str, start_function: &str) -> u32 {
    unsafe { raw::plug_function(method.as_bytes().as_ptr(), method.as_bytes().len() as u32, path.as_bytes().as_ptr(), path.as_bytes().len() as u32, name.as_bytes().as_ptr(), name.as_bytes().len() as u32, start_function.as_bytes().as_ptr(), start_function.as_bytes().len() as u32) }
}

pub fn plug_file(method: &str, path: &str, name: &str) -> u32 {
    unsafe { raw::plug_file(method.as_bytes().as_ptr(), method.as_bytes().len() as u32, path.as_bytes().as_ptr(), path.as_bytes().len() as u32, name.as_bytes().as_ptr(), name.as_bytes().len() as u32) }
}

pub fn unplug_path(method: &str, path: &str) -> u32 {
    unsafe { raw::unplug_path(method.as_bytes().as_ptr(), method.as_bytes().len() as u32, path.as_bytes().as_ptr(), path.as_bytes().len() as u32) }
}

pub fn get_status() -> Result<String, u32> {
    let result_buffer_id = unsafe { raw::get_status() };
    if result_buffer_id == 0xffff {
        Err(3)
    }
    else {
        let result_buffer = read_exchange_buffer(result_buffer_id);
        match result_buffer {
            Ok(result_buffer) => {
                let result = String::from_utf8(result_buffer).unwrap();
                //free_buffer(result_buffer_id);
                Ok(result)
            },
            Err(err) => {
                Err(4)
            },
        }
    }
}

pub fn persistence_set(key: &[u8], value: &[u8]) -> u32 {
    unsafe { raw::persistence_set(key.as_ptr(), key.len() as u32, value.as_ptr(), value.len() as u32) }
}

pub fn get_url(url: &str) -> Result<Vec<u8>, u32> {
    let result_buffer_id = unsafe { raw::get_url(url.as_bytes().as_ptr(), url.as_bytes().len() as u32) };
    if result_buffer_id == 0xffff {
        Err(1)
    }
    else {
        let result = read_exchange_buffer(result_buffer_id);
        match result {
            Ok(result) => {
                //free_buffer(result_buffer_id);
                Ok(result)
            },
            Err(err) => {
                Err(2)
            },
        }
    }
}

pub fn persistence_get(key: &[u8]) -> Result<Vec<u8>, u32> {
    let result_buffer_id = unsafe { raw::persistence_get(key.as_ptr(), key.len() as u32) };
    if result_buffer_id == 0xffff {
        Err(1)
    }
    else {
        let result = read_exchange_buffer(result_buffer_id);
        match result {
            Ok(result) => {
                //free_buffer(result_buffer_id);
                Ok(result)
            },
            Err(err) => {
                Err(2)
            },
        }
    }
}

pub fn persistence_get_subset(prefix: &str) -> Result<HashMap<String,String>, u32> {
    let result_buffer_id = unsafe { raw::persistence_get_subset(prefix.as_bytes().as_ptr(), prefix.as_bytes().len() as u32) };
    if result_buffer_id == 0xffff {
        Err(5)
    }
    else {
        let result_buffer = read_exchange_buffer(result_buffer_id);
        match result_buffer {
            Ok(result_buffer) => {
                let mut result = HashMap::new();
                let buffer = result_buffer.as_slice();
                free_buffer(result_buffer_id);

                let mut rdr = Cursor::new(buffer);
                let mut current_key : String = "".to_string();

                let mut nb_buffers = rdr.read_u32::<LittleEndian>().unwrap();
                while nb_buffers > 0 {
                    let buffer_size = rdr.read_u32::<LittleEndian>().unwrap();
                    
                    let mut buffer = Vec::with_capacity(buffer_size as usize);
                    unsafe { buffer.set_len(buffer_size as usize); }
                    rdr.read(&mut buffer);

                    let s = String::from_utf8(buffer.to_vec()).unwrap();
                    
                    if nb_buffers % 2 == 0 {
                        current_key = s;
                    } else {
                        result.insert(current_key.clone(), s);
                    }

                    nb_buffers = nb_buffers - 1;        
                }

                Ok(result)
            },
            Err(err) => {
                Err(6)
            },
        }
    }
}

pub fn print_debug(text: &str) -> u32 {
    unsafe { raw::print_debug(text.as_bytes().as_ptr(), text.as_bytes().len() as u32) }
}

pub fn get_time(dest: &[u8]) -> u32 {
    unsafe { raw::get_time(dest.as_ptr(), dest.len() as u32) }
}

pub fn free_buffer(bufferId:u32) -> u32 {
    unsafe { raw::free_buffer(bufferId) }
}

pub fn call_function(name: &str, start_function: &str, arguments: &[u32], mode: &str, input_exchange_buffer_id:u32, output_exchange_buffer_id:u32, posix_file_name: &str, posix_arguments: &[&str]) -> u32 {
    unsafe { raw::call_function(name.as_bytes().as_ptr(), name.as_bytes().len() as u32, start_function.as_bytes().as_ptr(), start_function.as_bytes().len() as u32, arguments.as_ptr(), arguments.len() as u32, mode.as_bytes().as_ptr(), mode.as_bytes().len() as u32, input_exchange_buffer_id, output_exchange_buffer_id, posix_file_name.as_bytes().as_ptr(), posix_file_name.as_bytes().len() as u32, std::ptr::null(), 0) }
}

