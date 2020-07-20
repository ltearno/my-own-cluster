MOC := ../../my-own-cluster

all: reverseproxy

.PHONY: reverseproxy
reverseproxy: reverseproxy.js core-api-guest.d.ts
	@echo pushing files
	@$(MOC) push reverseproxy reverseproxy.js
	@$(MOC) plug '/reverseproxy/*path' reverseproxy invoke 'http://server:8081'
	@echo you can go to https://localhost:8443/reverseproxy/ in your browser

core-api-guest.d.ts:
	$(MOC) guest-api core ts > core-api-guest.d.ts

clean:
	rm -f core-api-guest.d.ts