all: uppercase.wasm wasi-write.wasm hello-go.wasm http-request.wasm api-demo-c.wasm

http-request.wasm: http-request.c
	wasicc http-request.c -o http-request.wasm

uppercase.wasm: uppercase.c
	wasicc uppercase.c -o uppercase.wasm

api-demo-c.wasm: api-demo-c.c
	wasicc  -Os                                        \
        -z stack-size=4096 -Wl,--initial-memory=65536  \
        -Wl,--allow-undefined-file=api-demo-c.syms     \
        -Wl,--strip-all -nostdlib                      \
        -o api-demo-c.wasm api-demo-c.c

	# Optimize (optional)
	#wasm-opt -O3 api-demo-c.wasm -o api-demo-c.wasm
	#wasm-strip api-demo-c.wasm

	wasm2wat api-demo-c.wasm > api-demo-c.wat

wasi-write.wasm: wasi-write.wat
	wat2wasm wasi-write.wat -o wasi-write.wasm

hello-go.wasm: hello-go.go
	#GOOS=js GOARCH=wasm go build -o hello-go.wasm hello-go.go
	tinygo build -o hello-go.wasm -target wasm -wasm-abi generic hello-go.go
	wasm2wat hello-go.wasm > hello-go.wat

# you should have the rust toolchain to build that
wasm-rust-demo/target/wasm32-wasi/debug/wasm-rust-demo.wasm: wasm-rust-demo/src/main.rs
	@echo Compiling rust program to webassembly with wasi
	@cd wasm-rust-demo && rustup target add wasm32-wasi && cargo build --target wasm32-wasi


.PHONY: register register-rust call

register: http-request.wasm api-demo-c.wasm uppercase.wasm wasi-write.wasm hello-go.wasm
	@echo Registering functions...
	@../my-own-cluster push http-request http-request.wasm
	@../my-own-cluster push uppercase uppercase.wasm
	@../my-own-cluster push api-demo-c api-demo-c.wasm
	@../my-own-cluster push wasi-write wasi-write.wasm
	@../my-own-cluster push hello-go hello-go.wasm
	@../my-own-cluster push cowsay cowsay.wasm

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
	@#curl -k -s https://localhost:8443/api/function/call -H 'Content-Type: application/json' -d '{"mode":"posix","name":"cowsay","arguments":["I wish I was a cow !"],"wasi_file_name":"cowsay"}' | jq
	@#echo

register-rust: wasm-rust-demo/target/wasm32-wasi/debug/wasm-rust-demo.wasm
	@echo Registering rust sample
	@../my-own-cluster push wasm-rust-demo wasm-rust-demo/target/wasm32-wasi/debug/wasm-rust-demo.wasm

call-rust:
	@echo Calling 'wasm-rust-demo', a little demo program written in Rust and built for the wasm32-wasi platform
	../my-own-cluster call wasm-rust-demo -mode posix "https://home.lteconsulting.fr" "api://output" | jq
	@echo Decoding base64 output :
	../my-own-cluster call wasm-rust-demo -mode posix "https://home.lteconsulting.fr" "api://output" | jq -r '.output' | base64 -d
	@echo

test-files:
	@echo upload a file to some path
	@../my-own-cluster upload /titi/uppercase.c uppercase.c
	@echo curl the path :
	@curl -k https://localhost:8443/titi/uppercase.c