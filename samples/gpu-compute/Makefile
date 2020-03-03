MOC := ../../my-own-cluster

all: register

.PHONY: register
register:
	@echo pushing files
	@$(MOC) push gpu-compute-shader shader.glsl
	@$(MOC) push gpu-compute application.js
	@$(MOC) call gpu-compute -start_function test
	#@$(MOC) call gpu-compute -start_function launchMandelbrotShader | jq -r '.output' | base64 -d | od -F
	@$(MOC) plug /gpu-compute/mandelbrot gpu-compute launchMandelbrotShader
	@$(MOC) upload /gpu-compute/ index.html
	@echo "To create a test.png file with mandelbrot, you can run curl -k https://localhost:8443/gpu-compute/mandelbrot > test.png"