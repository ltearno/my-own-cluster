// When the `wee_alloc` feature is enabled, use `wee_alloc` as the global
// allocator.
#[cfg(feature = "wee_alloc")]
#[global_allocator]
static ALLOC: wee_alloc::WeeAlloc = wee_alloc::WeeAlloc::INIT;

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
