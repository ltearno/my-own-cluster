MOC := ../../my-own-cluster

all: aaa

.PHONY: aaa
aaa: core-api-guest.d.ts
	@echo pushing files
	@$(MOC) push aaa aaa.js

.PHONY: set-filter
set-filter:
	$(MOC) plug-filter aaa filter

core-api-guest.d.ts: $(MOC)
	$(MOC) guest-api core ts > core-api-guest.d.ts

clean:
	rm -f core-api-guest.d.ts