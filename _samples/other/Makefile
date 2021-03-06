MOC := ../../my-own-cluster
COMPILED_TARGETS := http-request.wasm api-demo-c.wasm uppercase.wasm wasi-write.wasm hello-go.wasm interop.wasm api-demo-c.wasm

all: http-request uppercase api-demo-c wasi-write hello-go

http-request.wasm: http-request.c
	wasicc http-request.c -o http-request.wasm

uppercase.wasm: uppercase.c
	wasicc uppercase.c -o uppercase.wasm

api-demo-c.wasm: api-demo-c.c
	# generate api stubs
	$(MOC) guest-api core c > api-demo-guest-api.h
	$(MOC) guest-api core c-syms > api-demo-guest-api.syms
	# compile
	wasicc  -Os                                             \
        -z stack-size=4096 -Wl,--initial-memory=65536       \
        -Wl,--allow-undefined-file=api-demo-guest-api.syms  \
        -Wl,--strip-all -Wl,--no-entry -nostdlib            \
        -o api-demo-c.wasm api-demo-c.c
	# Optimize (optional)
	#wasm-opt -O3 api-demo-c.wasm -o api-demo-c.wasm
	wasm-strip api-demo-c.wasm
	# dump wasm text format
	wasm2wat api-demo-c.wasm > api-demo-c.wat

interop.wasm: interop.c
	# generate api stubs
	$(MOC) guest-api core c > interop-guest-api.h
	$(MOC) guest-api core c-syms > interop-guest-api.syms
	# add imported functions to wasicc linker white list
	echo "get_size_of_passed_buffer" >> interop-guest-api.syms
	echo "rustDivide" >> interop-guest-api.syms
	echo "rustMultiply" >> interop-guest-api.syms
	# compile
	wasicc  -Os                                             \
        -z stack-size=4096 -Wl,--initial-memory=65536       \
        -Wl,--allow-undefined-file=interop-guest-api.syms   \
        -Wl,--strip-all -Wl,--no-entry -nostdlib            \
        -o interop.wasm interop.c

wasi-write.wasm: wasi-write.wat
	wat2wasm wasi-write.wat -o wasi-write.wasm

hello-go.wasm: hello-go.go
	#GOOS=js GOARCH=wasm go build -o hello-go.wasm hello-go.go
	tinygo build -o hello-go.wasm -target wasm -wasm-abi generic hello-go.go
	wasm2wat hello-go.wasm > hello-go.wat

.PHONY: clean
clean:
	rm $(COMPILED_TARGETS)

.PHONY: http-request
http-request: http-request.wasm
	@$(MOC) push http-request http-request.wasm
	@echo Calling 'http-request' POSIX mode, the client code makes a http request through the my-own-cluster api and outputs its reply body
	@$(MOC) call http-request -mode posix -input "Hello to you application" https://lkml.org/ | jq -r '.output' | base64 -d
	@echo

.PHONY: uppercase
uppercase: uppercase.wasm
	@$(MOC) push uppercase uppercase.wasm
	@echo Calling 'uppercase' POSIX mode, result is the uppercase of the buffer given in the http request
	@$(MOC) call uppercase -mode posix -input "This text should be in uppercase, if yes, you can say yeahh !" | jq -r '.output' | base64 -d
	@echo

.PHONY: api-demo-c
api-demo-c: api-demo-c.wasm
	@$(MOC) push api-demo-c api-demo-c.wasm
	@echo Calling 'api-demo-c' DIRECT mode, result is the addition of the two parameters we pass in the http query
	@$(MOC) call api-demo-c 264 55 | jq
	@echo

.PHONY: wasi-write
wasi-write: wasi-write.wasm
	@$(MOC) push wasi-write wasi-write.wasm
	@echo Calling 'wasi-write' POSIX mode, a simple webassembly text compiled program
	@$(MOC) call wasi-write -mode posix -wasi_file_name some_name_eh_eh | jq -r '.output' | base64 -d
	@echo

.PHONY: hello-go
hello-go: hello-go.wasm
	@$(MOC) push hello-go hello-go.wasm
	@echo Calling 'hello-go' POSIX mode, a golang program compiled with TinyGo
	$(MOC) call hello-go -mode posix
	@echo

.PHONY: interop
interop: interop.wasm
	@$(MOC) push interop interop.wasm
	# TODO move await rustMultiply function from watchdog !!!
	@echo Calling multiply function implemented in Rust without wasi, result of the multiplication is found in the result field
	$(MOC) call rust-multiply -start_function rustMultiply 4 3 | jq
	@echo
	@echo Calling 'interop' DIRECT mode, result is processed in the C code that calls Rust code
	$(MOC) call interop -start_function process 264 55 | jq
	@echo

test-files:
	@echo upload a file to some path
	@$(MOC) upload /titi/uppercase.c uppercase.c
	@echo curl the path :
	@curl -k https://localhost:8443/titi/uppercase.c