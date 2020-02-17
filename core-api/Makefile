TARGET = target/wasm32-unknown-unknown/debug/my_own_cluster_core_api.wasm

all: register

$(TARGET): src/lib.rs src/moc.rs
	@echo Building my-own-cluster-core-api module
	@rustup target add wasm32-unknown-unknown
	@cargo build --target=wasm32-unknown-unknown

register: $(TARGET)
	@my-own-cluster push core-api $(TARGET)
	@my-own-cluster plug -method post /my-own-cluster/api/file/register core-api plugFile