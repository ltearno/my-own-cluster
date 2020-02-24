use std::collections::HashMap;
use std::io::{Cursor, Read};
use byteorder::{LittleEndian, ReadBytesExt};

/**
 * my-own-cluster guest API bindings for rust
 * 
 * It can be called from rust code compiled into wasm and executed by my-own-cluster
 */

// import the 'my-own-cluster' module
pub mod raw {
    #[link(wasm_import_module = "my-own-cluster")]
    extern {
        pub fn test() -> u32;
        pub fn print_debug(buffer: *const u8, length: u32) -> u32;
        pub fn get_time(timestamp: *mut i64) -> u32;
        pub fn register_buffer(buffer: *const u8, length: u32) -> u32;
        pub fn get_buffer_size(buffer_id: u32) -> u32;
        pub fn get_buffer(buffer_id: u32, buffer: *const u8, length: u32) -> u32;
        // returns the exchange buffer id where headers have been copied or 0xffff if error
        pub fn get_buffer_headers(buffer_id: u32) -> u32;
        pub fn write_buffer(buffer_id: u32, buffer: *const u8, length: u32) -> u32;
        pub fn write_buffer_header(buffer_id: u32, name: *const u8, name_length: u32, value: *const u8, value_length: u32) -> u32;
        pub fn free_buffer(buffer_id: u32) -> i32;
        pub fn get_input_buffer_id() -> u32;
        pub fn get_output_buffer_id() -> u32;
        pub fn get_url(buffer: *const u8, length: u32) -> u32;
        pub fn persistence_set(key_buffer: *const u8, key_length: u32, value_buffer: *const u8, value_length: u32) -> u32;
        // returns the exchange buffer id or 0xffff if error
        pub fn persistence_get(key_buffer: *const u8, key_length: u32) -> u32;
        pub fn persistence_get_subset(prefix_buffer: *const u8, prefix_length: u32) -> u32;

        pub fn plug_file(method_buffer: *const u8, method_length: u32, path_buffer: *const u8, path_length: u32, content_type_buffer: *const u8, content_type_length: u32, bytes_buffer: *const u8, bytes_length: u32) -> u32;

        // temporary
        pub fn base64_decode(encoded_buffer: *const u8, encoded_length: u32) -> u32;
    }
}

pub fn base64_decode(encoded: &str) -> Result<Vec<u8>, &str> {
    let decoded_buffer_id = unsafe { raw::base64_decode(encoded.as_ptr(), encoded.len() as u32) };
    if decoded_buffer_id == 0xffff {
        return Err("cannot decode")
    }

    let decoded = read_buffer(decoded_buffer_id);
    free_buffer(decoded_buffer_id);

    Ok(decoded)
}

pub fn plug_file(method: &str, path: &str, content_type: &str, bytes: &[u8]) -> String {
    let method = method.as_bytes();
    let path = path.as_bytes();
    let content_type = content_type.as_bytes();

    let tech_id_buffer_id = unsafe { raw::plug_file(method.as_ptr(), method.len() as u32, path.as_ptr(), path.len() as u32, content_type.as_ptr(), content_type.len() as u32, bytes.as_ptr(), bytes.len() as u32) };
    let tech_id = read_buffer_as_string(tech_id_buffer_id);
    free_buffer(tech_id_buffer_id);

    tech_id
}

pub fn persistence_set(key :&str, value:&str) {
    unsafe {
        let key = key.as_bytes();
        let value = value.as_bytes();

        raw::persistence_set(key.as_ptr(), key.len() as u32, value.as_ptr(), value.len() as u32);
    }
}

pub fn persistence_get(key : &str) -> Option<String> {
    unsafe {
        let key = key.as_bytes();

        let buffer_id = raw::persistence_get(key.as_ptr(), key.len() as u32);
        if buffer_id == 0xffff {
            None
        }
        else {
            let value = read_buffer_as_string(buffer_id);
            free_buffer(buffer_id);

            Some(value)
        }
    }
}

pub fn get_buffer_headers(buffer_id: u32) -> HashMap<String,String> {
    let mut result = HashMap::new();
    
    let buffer_id = unsafe { raw::get_buffer_headers(buffer_id) };
    if buffer_id == 0xffff {
        return result;
    }

    let b = read_buffer(buffer_id);
    let buffer = b.as_slice();
    free_buffer(buffer_id);

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

    result
}

pub fn persistence_get_subset(prefix : &str) -> HashMap<String,String> {
    let mut result = HashMap::new();

    let prefix = prefix.as_bytes();
    
    let buffer_id = unsafe { raw::persistence_get_subset(prefix.as_ptr(), prefix.len() as u32) };
    if buffer_id == 0xffff {
        return result;
    }

    let b = read_buffer(buffer_id);
    let buffer = b.as_slice();
    free_buffer(buffer_id);

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

    result
}

pub fn get_input_buffer_id() -> u32 {
    unsafe {
        raw::get_input_buffer_id()
    }
}

pub fn get_output_buffer_id() -> u32 {
    unsafe {
        raw::get_output_buffer_id()
    }
}

pub fn print_debug(s: &str) {
    let bytes = s.as_bytes();
    unsafe {        
        raw::print_debug(bytes.as_ptr(), bytes.len() as u32);
    }
}

pub fn get_time() -> i64 {
    let mut timestamp : i64 = 0;

    unsafe {
        raw::get_time(&mut timestamp);
    }

    timestamp
}

pub fn free_buffer(buffer_id: u32) {
    unsafe {
        raw::free_buffer(buffer_id);
    }
}

pub fn read_buffer(buffer_id: u32) -> Vec<u8> {
    unsafe {
        let buffer_size = raw::get_buffer_size(buffer_id);
        let mut dest = Vec::with_capacity(buffer_size as usize);
        dest.resize(buffer_size as usize, 0);
        raw::get_buffer(buffer_id, dest.as_ptr(), buffer_size);
        dest
    }
}

pub fn read_buffer_as_string(buffer_id: u32) -> String {
    String::from_utf8(read_buffer(buffer_id)).unwrap()
}

pub fn get_url(url: &str) -> String {
    unsafe {
        let bytes = url.as_bytes();
        let buffer_id = raw::get_url(bytes.as_ptr(), bytes.len() as u32);
        let content = read_buffer_as_string(buffer_id);
        free_buffer(buffer_id);

        content
    }
}

pub fn write_buffer(buffer_id: u32, buffer: &[u8]) -> u32 {
    unsafe {
        raw::write_buffer(buffer_id, buffer.as_ptr(), buffer.len() as u32)
    }
}

pub fn write_buffer_header(buffer_id: u32, name: &str, value: &str) {
    unsafe {
        raw::write_buffer_header(
            buffer_id, 
            name.as_bytes().as_ptr(), 
            name.as_bytes().len() as u32, 
            value.as_bytes().as_ptr(), 
            value.as_bytes().len() as u32);
    }
}