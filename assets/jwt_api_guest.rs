
use std::collections::HashMap;
use std::io::{Cursor, Read};
use byteorder::{LittleEndian, ReadBytesExt};

/**
 * 'jwt' guest API bindings for rust
 * 
 * it can be called from rust code compiled into wasm and executed by my-own-cluster
 * 
 * you can use it by adding that in your rust source files at the beginning :
 * 
 *  mod jwt_api_guest;
 *  use jwt_api_guest::*;
 */

// import the 'jwt' module
pub mod raw {
    #[link(wasm_import_module = "jwt")]
    extern {
        pub fn verify_jwt(jwt_string: *const u8, jwt_length: u32) -> u32;

    }
}

pub fn verify_jwt(jwt: &str) -> Result<String, u32> {
    let result_buffer_id = unsafe { raw::verify_jwt(jwt.as_bytes().as_ptr(), jwt.as_bytes().len() as u32) };
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

