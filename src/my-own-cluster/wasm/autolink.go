package wasm

type AutoLinkAPIPlugin struct{}

func NewAutoLinkAPIPlugin() APIPlugin {
	return &AutoLinkAPIPlugin{}
}

func (p *AutoLinkAPIPlugin) Bind(wctx *WasmProcessContext) {
}
