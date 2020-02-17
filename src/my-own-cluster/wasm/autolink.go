package wasm

type AutoLinkWASMAPIPlugin struct{}

func NewAutoLinkWASMAPIPlugin() WASMAPIPlugin {
	return &AutoLinkWASMAPIPlugin{}
}

func (p *AutoLinkWASMAPIPlugin) Bind(wctx *WasmProcessContext) {
}
