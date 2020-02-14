SERVER_PUBLIC_URL ?= https://localhost:8443
MYOWNCLUSTER_SERVER_BASE_URL ?= https://localhost:8444

all: uppercase.wasm wasi-write.wasm hello-go.wasm http-request.wasm api-demo-c.wasm

http-request.wasm: http-request.c
	wasicc http-request.c -o http-request.wasm

uppercase.wasm: uppercase.c
	wasicc uppercase.c -o uppercase.wasm

api-demo-c.wasm: api-demo-c.c api-demo-c.h api-demo-c.syms
	wasicc  -Os                                        \
        -z stack-size=4096 -Wl,--initial-memory=65536  \
        -Wl,--allow-undefined-file=api-demo-c.syms     \
        -Wl,--strip-all -Wl,--no-entry -nostdlib                      \
        -o api-demo-c.wasm api-demo-c.c

	# Optimize (optional)
	#wasm-opt -O3 api-demo-c.wasm -o api-demo-c.wasm
	#wasm-strip api-demo-c.wasm

	wasm2wat api-demo-c.wasm > api-demo-c.wat

interop.wasm: interop.c api-demo-c.h
	wasicc  -Os                                        \
        -z stack-size=4096 -Wl,--initial-memory=65536  \
        -Wl,--allow-undefined-file=interop.syms     \
        -Wl,--strip-all -Wl,--no-entry -nostdlib                      \
        -o interop.wasm interop.c

wasi-write.wasm: wasi-write.wat
	wat2wasm wasi-write.wat -o wasi-write.wasm

hello-go.wasm: hello-go.go
	#GOOS=js GOARCH=wasm go build -o hello-go.wasm hello-go.go
	tinygo build -o hello-go.wasm -target wasm -wasm-abi generic hello-go.go
	wasm2wat hello-go.wasm > hello-go.wat

# you should have the rust toolchain to build that
# some starting info here : https://github.com/bytecodealliance/wasmtime/blob/master/docs/WASI-tutorial.md#from-rust
wasm-rust-demo/target/wasm32-wasi/debug/wasm-rust-demo.wasm: wasm-rust-demo/src/main.rs
	@echo Compiling rust program to webassembly with wasi
	@cd wasm-rust-demo && rustup target add wasm32-wasi && cargo build --target wasm32-wasi

rust-nowasi/target/wasm32-unknown-unknown/debug/rust_nowasi.wasm: rust-nowasi/src/lib.rs rust-nowasi/src/moc.rs
	@echo Building rust program without wasi
	@cd rust-nowasi && rustup target add wasm32-unknown-unknown && cargo build --target=wasm32-unknown-unknown

.PHONY: register register-rust call

register: http-request.wasm api-demo-c.wasm uppercase.wasm wasi-write.wasm hello-go.wasm interop.wasm api-demo-c.wasm
	@echo Registering functions...
	@../my-own-cluster push http-request http-request.wasm
	@../my-own-cluster push uppercase uppercase.wasm
	@../my-own-cluster push api-demo-c api-demo-c.wasm
	@../my-own-cluster push wasi-write wasi-write.wasm
	@../my-own-cluster push hello-go hello-go.wasm
	@../my-own-cluster push cowsay cowsay.wasm
	@../my-own-cluster push interop interop.wasm

call:
	@echo Calling 'http-request' POSIX mode, the client code makes a http request through the my-own-cluster api and outputs its reply body
	@../my-own-cluster call http-request -mode posix -input "Hello to you application" https://home.lteconsulting.fr | jq -r '.output' | base64 -d
	@echo

	@echo Calling 'uppercase' POSIX mode, result is the uppercase of the buffer given in the http request
	@../my-own-cluster call uppercase -mode posix -input "This text should be in uppercase, if yes, you can say yeahh !" | jq -r '.output' | base64 -d
	@echo
	@echo

	@echo Calling 'api-demo-c' DIRECT mode, result is the addition of the two parameters we pass in the http query
	@../my-own-cluster call api-demo-c 264 55 | jq
	@echo

	@echo Calling 'wasi-write' POSIX mode, a simple webassembly text compiled program
	@../my-own-cluster call wasi-write -mode posix -wasi_file_name some_name_eh_eh | jq -r '.output' | base64 -d
	@echo

	@echo Calling 'hello-go' POSIX mode, a golang program compiled with TinyGo
	../my-own-cluster call hello-go -mode posix
	@echo

	@#echo Calling 'cowsay' POSIX mode, does not work yet !
	@#curl -k -s $(SERVER_PUBLIC_URL)/api/function/call -H 'Content-Type: application/json' -d '{"mode":"posix","name":"cowsay","arguments":["I wish I was a cow !"],"wasi_file_name":"cowsay"}' | jq
	@#echo

register-rust: wasm-rust-demo/target/wasm32-wasi/debug/wasm-rust-demo.wasm rust-nowasi/target/wasm32-unknown-unknown/debug/rust_nowasi.wasm
	@echo Registering rust sample
	@../my-own-cluster push wasm-rust-demo wasm-rust-demo/target/wasm32-wasi/debug/wasm-rust-demo.wasm
	@../my-own-cluster push rust-multiply rust-nowasi/target/wasm32-unknown-unknown/debug/rust_nowasi.wasm

	@../my-own-cluster upload /watchdog/ rust-nowasi/index.html
	@../my-own-cluster push watchdog-backend rust-nowasi/target/wasm32-unknown-unknown/debug/rust_nowasi.wasm
	@../my-own-cluster plug /watchdog/status watchdog-backend getStatus
	@../my-own-cluster plug -method post /watchdog/status watchdog-backend postStatus

call-rust:
	@echo Calling 'wasm-rust-demo', a little demo program written in Rust and built for the wasm32-wasi platform
	../my-own-cluster call wasm-rust-demo -mode posix "https://home.lteconsulting.fr" "api://output" | jq
	@echo Decoding base64 output :
	../my-own-cluster call wasm-rust-demo -mode posix "https://home.lteconsulting.fr" "api://output" | jq -r '.output' | base64 -d
	@echo
	@echo Calling multiply function implemented in Rust without wasi, result of the multiplication is found in the result field
	../my-own-cluster call rust-multiply -start_function rustMultiply 4 3 | jq
	curl -k $(SERVER_PUBLIC_URL)/watchdog/
	#my-own-cluster call watchdog-backend -start_function getStatus | jq -r '.output' | base64 -d | jq
	curl -s -k $(SERVER_PUBLIC_URL)/watchdog/status | jq
	curl -s -k -X POST $(SERVER_PUBLIC_URL)/watchdog/status -H "content-type: application/json" -d '{"name":"backup-daemon","message":"all data backuped"}'
	sleep 1
	curl -s -k -X POST $(SERVER_PUBLIC_URL)/watchdog/status -H "content-type: application/json" -d '{"name":"home-light-1","message":"everything ok"}'

call-interop:
	@echo Calling 'interop' DIRECT mode, result is processed in the C code that calls Rust code
	../my-own-cluster call interop -start_function process 264 55 | jq
	@echo

test-files:
	@echo upload a file to some path
	@../my-own-cluster upload /titi/uppercase.c uppercase.c
	@echo curl the path :
	@curl -k $(SERVER_PUBLIC_URL)/titi/uppercase.c