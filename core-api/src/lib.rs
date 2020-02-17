use serde::{Deserialize, Serialize};
use std::str;

mod moc;
use moc::*;

// When the `wee_alloc` feature is enabled, use `wee_alloc` as the global
// allocator.
#[cfg(feature = "wee_alloc")]
#[global_allocator]
static ALLOC: wee_alloc::WeeAlloc = wee_alloc::WeeAlloc::INIT;

#[derive(Serialize)]
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

#[derive(Deserialize)]
struct PlugFileRequest {
	method: String,
	path:String,
	content_type: String,
	bytes: String,
}

#[derive(Serialize)]
struct PlugFileResponse {
	status:      bool,
	tech_id:      String,
	method:      String,
	path:        String,
	content_type: String,
	bytes_size:   u32,
}

#[no_mangle]
pub extern fn plugFile() -> u32 {
    let body = read_buffer_as_string(get_input_buffer_id());
    match serde_json::from_str::<PlugFileRequest>(&body) {
        Ok(request) => {
            match base64_decode(&request.bytes) {
                Ok(bytes) => {
                    let tech_id = plug_file(&request.method, &request.path, &request.content_type, &bytes);

                    let response_body = &PlugFileResponse {
                        status: true,
                        tech_id: tech_id,
                        method: request.method,
                        path: request.path,
                        content_type: request.content_type,
                        bytes_size: bytes.len() as u32,
                    };
                
                    write_buffer_header(get_output_buffer_id(), "content-type", "application/json");
                    write_buffer(get_output_buffer_id(), serde_json::to_string(&response_body).unwrap().as_bytes());
                    200 as u32
                },

                Err(err) => {
                    message_response("cannot decode base64 bytes in request body");
                    400 as u32
                },
            }
        },
        Err(err) => {
            message_response("cannot parse request body");
            400 as u32
        },
    }
}
