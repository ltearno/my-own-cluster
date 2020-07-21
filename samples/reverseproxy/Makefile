MOC := ../../my-own-cluster

all: reverseproxy

.PHONY: reverseproxy
reverseproxy: reverseproxy.js core-api-guest.d.ts
	@echo pushing files
	@$(MOC) upload /test-ws index.html
	@$(MOC) push reverseproxy reverseproxy.js
	#@$(MOC) plug '/reverseproxy/ws' reverseproxy invoke '{"backend":"wss://home.lteconsulting.fr/blockchain-js/events"}'
	@$(MOC) plug '/reverseproxy/ws' reverseproxy invoke '{"backend":"ws://server:9091/events"}'
	@$(MOC) plug '/reverseproxy/star/*path' reverseproxy invoke '{"backend":"http://server:8081"}'
	@echo you can go to https://localhost:8443/reverseproxy/ in your browser

core-api-guest.d.ts:
	$(MOC) guest-api core ts > core-api-guest.d.ts

clean:
	rm -f core-api-guest.d.ts