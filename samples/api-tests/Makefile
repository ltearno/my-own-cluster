MOC := ../../my-own-cluster

all: api-test

.PHONY: api-test
api-test: api-test.js core-api-guest.d.ts
	@echo pushing files
	@$(MOC) push api-test api-test.js
	@$(MOC) call api-test -start_function run | jq -r '.output' | base64 -d | jq

core-api-guest.d.ts:
	$(MOC) guest-api core ts > core-api-guest.d.ts

clean:
	rm -f core-api-guest.d.ts