use serde::{Deserialize, Serialize};
use std::collections::HashMap;

// When the `wee_alloc` feature is enabled, use `wee_alloc` as the global
// allocator.
#[cfg(feature = "wee_alloc")]
#[global_allocator]
static ALLOC: wee_alloc::WeeAlloc = wee_alloc::WeeAlloc::INIT;

#[derive(Serialize, Deserialize)]
struct WatchdogStatus {
    id: i32,
    title: String,
}

// import the 'my-own-cluster' library
mod raw {
    #[link(wasm_import_module = "my-own-cluster")]
    extern {
        pub fn test() -> u32;
        pub fn print_debug(buffer: *const u8, length: u32) -> u32;
        pub fn register_buffer(buffer: *const u8, length: u32) -> u32;
        pub fn get_buffer_size(buffer_id: u32) -> u32;
        pub fn get_buffer(buffer_id: u32, buffer: *const u8, length: u32) -> u32;
        pub fn write_buffer(buffer_id: u32, buffer: *const u8, length: u32) -> u32;
        pub fn write_buffer_header(buffer_id: u32, name: *const u8, name_length: u32, value: *const u8, value_length: u32) -> u32;
        pub fn free_buffer(buffer_id: u32) -> i32;
        pub fn get_input_buffer_id() -> u32;
        pub fn get_output_buffer_id() -> u32;
        pub fn get_url(buffer: *const u8, length: u32) -> u32;
    }
}

pub fn print_debug(s: &str) {
    unsafe {
        let bytes = s.as_bytes();
        
        unsafe {
            raw::print_debug(bytes.as_ptr(), bytes.len() as u32);
        }
    }
}

pub fn get_url(url: &str) -> String {
    unsafe {
        let bytes = url.as_bytes();
        let buffer_id = raw::get_url(bytes.as_ptr(), bytes.len() as u32);
        let buffer_size = raw::get_buffer_size(buffer_id);
        let mut dest = Vec::with_capacity(buffer_size as usize);
        dest.resize(buffer_size as usize, 0);
        raw::get_buffer(buffer_id, dest.as_ptr(), buffer_size);
        raw::free_buffer(buffer_id);
        String::from_utf8(dest).unwrap()
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

#[no_mangle]
pub extern fn getStatus() -> u32 {
    let r = get_url("https://jsonplaceholder.typicode.com/todos/1");

    let rStatus = serde_json::from_str::<WatchdogStatus>(&r);
    match rStatus {
        Ok(status) => {
            match serde_json::to_string(&status) {
                Ok(serialized) => {
                    unsafe {
                        write_buffer_header(raw::get_output_buffer_id(), "content-type", "application/json");
                        raw::write_buffer(raw::get_output_buffer_id(), serialized.as_bytes().as_ptr(), serialized.as_bytes().len() as u32)
                    };

                    print_debug(&format!("serialized our response which is {} bytes long", serialized.as_bytes().len()));
                    
                    serialized.as_bytes().len() as u32
                },
                Err(err) => {
                    print_debug(&format!("error when serializing our response : {}", err));
                    13 as u32
                },
            }
        },
        Err(err) => {
            print_debug(&format!("error when deserializing backend response : {}", err));
            11 as u32
        },
    }
}

// import the 'my-own-cluster' library
#[link(wasm_import_module = "my-own-cluster")]
extern {
    fn test();
}

#[no_mangle]
pub extern fn rustMultiply(a : i32, b:i32) -> i32 {
    unsafe {
        test();
    }

    a * b
}

#[no_mangle]
pub extern fn rustDivide(a : i32, b:i32) -> i32 {
    a / b
}
