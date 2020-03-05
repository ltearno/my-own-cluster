use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::str;

mod core_api_guest;
use core_api_guest::*;

// When the `wee_alloc` feature is enabled, use `wee_alloc` as the global
// allocator.
#[cfg(feature = "wee_alloc")]
#[global_allocator]
static ALLOC: wee_alloc::WeeAlloc = wee_alloc::WeeAlloc::INIT;

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

    write_exchange_buffer_header(get_output_buffer_id(), "content-type", "application/json");
    write_exchange_buffer(get_output_buffer_id(), serialized.as_bytes());
}

#[no_mangle]
pub extern fn getStatus() -> u32 {
    let status = &mut WatchdogStatus{
        description: "everything ok".to_string(),
        services: HashMap::new(),
    };

    let entries = persistence_get_subset("/watchdog-v1/status/services/").unwrap();
    for (k, v) in entries.iter() {
        if k.ends_with("/timestamp") {
            let service_name = k.trim_start_matches("/watchdog-v1/status/services/").trim_end_matches("/timestamp").to_string();

            print_debug(&format!(" => service_name:{}", service_name));
            status.services.insert(service_name, WatchdogServiceStatus{ timestamp: v.parse().unwrap() });
        }
    }

    let serialized = serde_json::to_string(&status).unwrap();

    write_exchange_buffer_header(get_output_buffer_id(), "content-type", "application/json");
    write_exchange_buffer(get_output_buffer_id(), serialized.as_bytes());
    
    200
}

#[no_mangle]
pub extern fn addStatus() -> u32 {
    let headers = read_exchange_buffer_headers(get_input_buffer_id()).unwrap();
    let service_name = &headers["x-moc-path-param-service"];

    let timestampBytes = [0x56, 0x34, 0x12, 0x90, 0x78, 0x56, 0x34, 0x12];
    get_time(&timestampBytes);
    let mut timestamp : i64 = i64::from_le_bytes(timestampBytes) / 1000;
    persistence_set(&format!("/watchdog-v1/status/services/{}/timestamp", service_name).as_bytes(), &format!("{}", timestamp).as_bytes());

    message_response(&format!("status for '{}' saved for timestamp {}, thanks", service_name, timestamp));
    200 as u32
}

#[no_mangle]
pub extern fn postStatus() -> u32 {
    let body_buffer = read_exchange_buffer(get_input_buffer_id()).unwrap();
    let body = String::from_utf8(body_buffer).unwrap();
    let status = serde_json::from_str::<WatchdogServicePostStatus>(&body);
    match status {
        Ok(status) => {
            let timestampBytes = [0x56, 0x34, 0x12, 0x90, 0x78, 0x56, 0x34, 0x12];
            get_time(&timestampBytes);
            let mut timestamp : i64 = i64::from_le_bytes(timestampBytes) / 1000;
            persistence_set(&format!("/watchdog-v1/status/services/{}/timestamp", status.name).as_bytes(), &format!("{}", timestamp).as_bytes());

            message_response(&format!("status for '{}' saved for timestamp {}, thanks", status.name, timestamp));
            200 as u32
        },
        Err(err) =>{
            message_response(&format!("cannot parse {:?}", err));
            400 as u32
        },
    }
}

#[no_mangle]
pub extern fn rustMultiply(a : i32, b:i32) -> i32 {
    a * b
}

#[no_mangle]
pub extern fn rustDivide(a : i32, b:i32) -> i32 {
    a / b
}
