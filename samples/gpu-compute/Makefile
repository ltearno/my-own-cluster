MOC := ../../my-own-cluster

all: register

.PHONY: register
register:
	@echo pushing files
	@$(MOC) push gpu-compute-shader shader.glsl
	@$(MOC) plug /gpu-compute-shader/test-shader gpu-compute-shader
	@echo you can go to https://localhost:8443/gpu-compute-shader/test-shader in your browser to call the shader without input data
	@$(MOC) call gpu-compute-shader -input "Hello to you application" | jq