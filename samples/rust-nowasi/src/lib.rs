use serde::{Deserialize, Serialize};
use std::collections::HashMap;

mod moc;
use moc::*;

#[derive(Serialize, Deserialize)]
struct WatchdogServiceStatus {
    timestamp: i64,
}

#[derive(Serialize, Deserialize)]
struct WatchdogStatus {
    description: String,
    services: HashMap<String, WatchdogServiceStatus>,
}

#[derive(Serialize, Deserialize)]
struct WatchdogServicePostStatus {
    name: String,
    message: String,
}

#[derive(Serialize, Deserialize)]
struct MessageResponse {
    status: bool,
    message: String,
}

fn message_response(message: &str) {
    let body = &MessageResponse {
        status: true,
        message: message.to_string(),
    };

    let serialized = serde_json::to_string(&body).unwrap();

    write_buffer_header(get_output_buffer_id(), "content-type", "application/json");
    write_buffer(get_output_buffer_id(), serialized.as_bytes());
}

#[no_mangle]
pub extern fn getStatus() -> u32 {
    let status = &WatchdogStatus{
        description: "everything ok".to_string(),
        services: HashMap::new(),
    };

    let serialized = serde_json::to_string(&status).unwrap();

    write_buffer_header(get_output_buffer_id(), "content-type", "application/json");
    write_buffer(get_output_buffer_id(), serialized.as_bytes());
    
    200
}

#[no_mangle]
pub extern fn postStatus() -> u32 {
    let body = read_buffer_as_string(get_input_buffer_id());
    let status = serde_json::from_str::<WatchdogServicePostStatus>(&body);
    match status {
        Ok(status) => {
            // TODO
            // get timestamp
            // set key value for service in serverless database
            message_response(&format!("status for '{}' saved for timestamp {}, thanks", status.name, get_time()/1000));
            200 as u32
        },
        Err(err) =>{
            message_response("cannot parse");
            400 as u32
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
