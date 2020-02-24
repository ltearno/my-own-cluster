use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::str;

mod my_own_cluster_guest_api;
use my_own_cluster_guest_api::*;

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

    write_buffer_header(get_output_buffer_id(), "content-type", "application/json");
    write_buffer(get_output_buffer_id(), serialized.as_bytes());
}

#[no_mangle]
pub extern fn getStatus() -> u32 {
    let status = &mut WatchdogStatus{
        description: "everything ok".to_string(),
        services: HashMap::new(),
    };

    let entries = persistence_get_subset("/watchdog-v1/status/services/");
    for (k, v) in entries.iter() {
        if k.ends_with("/timestamp") {
            let service_name = k.trim_start_matches("/watchdog-v1/status/services/").trim_end_matches("/timestamp").to_string();

            print_debug(&format!(" => service_name:{}", service_name));
            status.services.insert(service_name, WatchdogServiceStatus{ timestamp: v.parse().unwrap() });
        }
    }

    let serialized = serde_json::to_string(&status).unwrap();

    write_buffer_header(get_output_buffer_id(), "content-type", "application/json");
    write_buffer(get_output_buffer_id(), serialized.as_bytes());
    
    200
}

#[no_mangle]
pub extern fn addStatus() -> u32 {
    let headers = get_buffer_headers(get_input_buffer_id());
    let service_name = &headers["x-moc-path-param-service"];

    persistence_set(&format!("/watchdog-v1/status/services/{}/timestamp", service_name), &format!("{}", get_time()/1000));

    message_response(&format!("status for '{}' saved for timestamp {}, thanks", service_name, get_time()/1000));
    200 as u32
}

#[no_mangle]
pub extern fn postStatus() -> u32 {
    let body = read_buffer_as_string(get_input_buffer_id());
    let status = serde_json::from_str::<WatchdogServicePostStatus>(&body);
    match status {
        Ok(status) => {
            persistence_set(&format!("/watchdog-v1/status/services/{}/timestamp", status.name), &format!("{}", get_time()/1000));

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
