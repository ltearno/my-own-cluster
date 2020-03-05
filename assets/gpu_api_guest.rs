
use std::collections::HashMap;
use std::io::{Cursor, Read};
use byteorder::{LittleEndian, ReadBytesExt};

/**
 * 'gpu' guest API bindings for rust
 * 
 * it can be called from rust code compiled into wasm and executed by my-own-cluster
 * 
 * you can use it by adding that in your rust source files at the beginning :
 * 
 *  mod gpu_api_guest;
 *  use gpu_api_guest::*;
 */

// import the 'gpu' module
pub mod raw {
    #[link(wasm_import_module = "gpu")]
    extern {
        pub fn compute_shader(specification_string: *const u8, specification_length: u32) -> u32;
        pub fn create_image_from_rgba_float_pixels(width:u32, height:u32, pixels_exchange_buffer_id:u32, png_exchange_buffer_id:u32) -> u32;
        pub fn create_image_from_r_float_pixels(width:u32, height:u32, pixels_exchange_buffer_id:u32, png_exchange_buffer_id:u32) -> u32;

    }
}

pub fn compute_shader(specification: &str) -> u32 {
    unsafe { raw::compute_shader(specification.as_bytes().as_ptr(), specification.as_bytes().len() as u32) }
}

pub fn create_image_from_rgba_float_pixels(width:u32, height:u32, pixels_exchange_buffer_id:u32, png_exchange_buffer_id:u32) -> u32 {
    unsafe { raw::create_image_from_rgba_float_pixels(width, height, pixels_exchange_buffer_id, png_exchange_buffer_id) }
}

pub fn create_image_from_r_float_pixels(width:u32, height:u32, pixels_exchange_buffer_id:u32, png_exchange_buffer_id:u32) -> u32 {
    unsafe { raw::create_image_from_r_float_pixels(width, height, pixels_exchange_buffer_id, png_exchange_buffer_id) }
}

