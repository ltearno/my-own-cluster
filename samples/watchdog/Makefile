MOC := ../../my-own-cluster
WASM_FILE := target/wasm32-unknown-unknown/debug/watchdog.wasm

all: register

$(WASM_FILE): src/lib.rs src/my_own_cluster_guest_api.rs
	@echo compiling rust program to webassembly
	@rustup target add wasm32-unknown-unknown
	@cargo build --target=wasm32-unknown-unknown

.PHONY: register
register: $(WASM_FILE)
	@echo pushing watchdog
	# associate url paths to files
	@$(MOC) upload /watchdog/ index.html
	@$(MOC) upload /watchdog/index.css surface_styles.css
	# register the wasm byte code and name it 'watchdog-backend'
	@$(MOC) push watchdog-backend $(WASM_FILE)
	# plug the GET '/watchdog/status' url path to the 'getStatus' function of our wasm code
	@$(MOC) plug /watchdog/status watchdog-backend getStatus
	# plug the GET '/watchdog/update/!service' url path to the 'addStatus' function of our wasm code
	@$(MOC) plug /watchdog/update/!service watchdog-backend addStatus
	# plug the POST '/watchdog/status' url path to the 'postStatus' function of our wasm code
	@$(MOC) plug -method post /watchdog/status watchdog-backend postStatus

	@echo you can now visit https://localhost:8443/watchdog/
	@echo to update a status, you can just curl -k -s https://localhost:8443/watchdog/update/SERVICE_NAME with any service name
	@echo you can also use the more REST idiomatic POST request : curl -s -k -X POST https://localhost:8443/watchdog/status -H "content-type: application/json" -d '{"name":"home-garage-lock","message":"locked, everything ok"}'

.PHONY: send-status
send-status:
	@echo sending some fake hearbeats
	@curl -k -s https://localhost:8443/watchdog/update/from-makefile | jq
	@sleep 1
	@curl -k -s https://localhost:8443/watchdog/update/another-one | jq

clean:
	rm $(WASM_FILE)