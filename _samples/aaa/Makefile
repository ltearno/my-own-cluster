MOC := ../../my-own-cluster

all: set-filter

.PHONY: set-filter
set-filter: core-api-guest.d.ts jwt-api-guest.d.ts
	@$(MOC) push aaa aaa.js
	$(MOC) plug-filter aaa filter

core-api-guest.d.ts: $(MOC)
	$(MOC) guest-api core ts > core-api-guest.d.ts

jwt-api-guest.d.ts: $(MOC)
	$(MOC) guest-api jwt ts > jwt-api-guest.d.ts

clean:
	rm -f core-api-guest.d.ts