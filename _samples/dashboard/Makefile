MOC := ../../my-own-cluster

all: dashboard

.PHONY: dashboard
dashboard: dashboard.js core-api-guest.d.ts
	@echo pushing files
	@$(MOC) push dashboard dashboard.js
	@$(MOC) push dashboard-template dashboard-template.html
	@$(MOC) plug /dashboard/ dashboard getDashboardHtml toto
	@$(MOC) upload /dashboard/index.css surface_styles.css
	@echo you can go to https://localhost:8443/dashboard/ in your browser

core-api-guest.d.ts:
	$(MOC) guest-api core ts > core-api-guest.d.ts

clean:
	rm -f core-api-guest.d.ts