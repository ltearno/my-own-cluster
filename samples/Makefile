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
	tinygo build -o hello-go.wasm -target wasm -wasm-abi generic hello-go.go
	wasm2wat hello-go.wasm > hello-go.wat


.PHONY: test-register test-call

test-register: http-request.wasm api-demo-c.wasm uppercase.wasm wasi-write.wasm hello-go.wasm
	@echo Registering functions...
	@echo Register 'http-request'
	@curl -k -s https://localhost:8443/api/functions/register -H 'Content-Type: application/json' -d '{"name":"http-request","wasm_bytes":"$(shell cat http-request.wasm | base64 -w 0)"}' | jq
	@echo Register 'uppercase'
	@curl -k -s https://localhost:8443/api/functions/register -H 'Content-Type: application/json' -d '{"name":"uppercase","wasm_bytes":"$(shell cat uppercase.wasm | base64 -w 0)"}' | jq
	@echo Register 'api-demo-c'
	@curl -k -s https://localhost:8443/api/functions/register -H 'Content-Type: application/json' -d '{"name":"api-demo-c","wasm_bytes":"$(shell cat api-demo-c.wasm | base64 -w 0)"}' | jq
	@echo Register 'wasi-write'
	@curl -k -s https://localhost:8443/api/functions/register -H 'Content-Type: application/json' -d '{"name":"wasi-write","wasm_bytes":"$(shell cat wasi-write.wasm | base64 -w 0)"}' | jq
	@echo Register 'hello-go'
	@curl -k -s https://localhost:8443/api/functions/register -H 'Content-Type: application/json' -d '{"name":"hello-go","wasm_bytes":"$(shell cat hello-go.wasm | base64 -w 0)"}' | jq
	@echo Register 'cowsay'
	@echo '{"name":"cowsay","wasm_bytes":"' > /tmp/tmp.json
	@cat /tmp/wapm_packages/_/cowsay@0.2.0/target/wasm32-wasi/release/cowsay.wasm | base64 -w 0 >> /tmp/tmp.json
	@echo '"}' >> /tmp/tmp.json
	@curl -k -s https://localhost:8443/api/functions/register -H "Content-Type: application/json" -H 'Content-Type: application/json' -d @/tmp/tmp.json | jq

test-call:
	@echo Calling 'http-request' POSIX mode, the client code makes a http request through the my-own-cluster api and outputs its reply body
	@curl -k -s https://localhost:8443/api/functions/call -H 'Content-Type: application/json' -d '{"mode":"posix","name":"http-request","arguments":["https://home.lteconsulting.fr"],"input":"Hello to you application","wasi_file_name":"http-request.wasm"}' | jq -r '.output' | base64 -d
	@echo

	@echo Calling 'uppercase' POSIX mode, result is the uppercase of the buffer given in the http request
	@curl -k -s https://localhost:8443/api/functions/call -H 'Content-Type: application/json' -d '{"mode":"posix","name":"uppercase","arguments":[],"input":"This text should be in uppercase\nIf yes, you can say yeahh !\n","wasi_file_name":"uppercase"}' | jq -r '.output' | base64 -d
	@echo

	@echo Calling 'api-demo-c' DIRECT mode, result is the addition of the two parameters we pass in the http query
	@curl -k -s https://localhost:8443/api/functions/call -H 'Content-Type: application/json' -d '{"mode":"direct","name":"api-demo-c","start_function":"_start","arguments":[2,5]}' | jq
	@echo

	@echo Calling 'wasi-write' POSIX mode, a simple webassembly text compiled program
	@curl -k -s https://localhost:8443/api/functions/call -H 'Content-Type: application/json' -d '{"mode":"posix","name":"wasi-write","arguments":[],"input":"This text should be in uppercase\nIf yes, you can say yeahh !\n","wasi_file_name":"uppercase"}' | jq -r '.output' | base64 -d
	@echo

	@echo Calling 'hello-go' POSIX mode, a golang program compiled with TinyGo
	@curl -k -s https://localhost:8443/api/functions/call -H 'Content-Type: application/json' -d '{"mode":"posix","name":"hello-go","arguments":[],"wasi_file_name":"uppercase"}' | jq
	@echo

	@#echo Calling 'cowsay' POSIX mode, does not work yet !
	@#curl -k -s https://localhost:8443/api/functions/call -H 'Content-Type: application/json' -d '{"mode":"posix","name":"cowsay","arguments":["I wish I was a cow !"],"wasi_file_name":"cowsay"}' | jq
	@#echo