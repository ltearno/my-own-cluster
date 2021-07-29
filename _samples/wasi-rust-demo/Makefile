WASM_FILE := target/wasm32-wasi/debug/wasi-rust-demo.wasm
MOC := ../../my-own-cluster

all: register call

# you should have the rust toolchain to build that
# some starting info here : https://github.com/bytecodealliance/wasmtime/blob/master/docs/WASI-tutorial.md#from-rust
$(WASM_FILE): src/main.rs
	@echo Compiling rust program to webassembly with wasi
	@rustup target add wasm32-wasi
	@cargo build --target wasm32-wasi

.PHONY: register
register: $(WASM_FILE)
	@echo registering rust sample
	@$(MOC) push wasi-rust-demo $(WASM_FILE)

.PHONY: call
call:
	@echo Calling 'wasi-rust-demo', a little demo program written in Rust and built for the wasm32-wasi platform
	$(MOC) call wasi-rust-demo -mode posix "https://lkml.org/" "api://output"
	@echo Decoding base64 output :
	$(MOC) call wasi-rust-demo -mode posix "https://lkml.org/" "api://output"