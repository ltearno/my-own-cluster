MOC := ../../my-own-cluster

all: register gpu-api-guest.d.ts core-api-guest.d.ts

.PHONY: register
register:
	@echo pushing files
	@$(MOC) push gpu-compute-shader shader.glsl
	@$(MOC) push gpu-compute application.js
	@$(MOC) plug /gpu-compute/mandelbrot gpu-compute launchMandelbrotShader
	@$(MOC) upload /gpu-compute/ index.html
	@$(MOC) call gpu-compute -start_function test
	#@$(MOC) call gpu-compute -start_function launchMandelbrotShader | jq -r '.output' | base64 -d | od -F

core-api-guest.d.ts:
	$(MOC) guest-api core ts > core-api-guest.d.ts

gpu-api-guest.d.ts:
	$(MOC) guest-api gpu ts > gpu-api-guest.d.ts