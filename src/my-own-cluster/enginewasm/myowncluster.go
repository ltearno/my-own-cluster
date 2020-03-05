package enginewasm

import (
	"fmt"
)

type MyOwnClusterWASMAPIPlugin struct{}

func NewMyOwnClusterWASMAPIPlugin() WASMAPIPlugin {
	return &MyOwnClusterWASMAPIPlugin{}
}

func (p *MyOwnClusterWASMAPIPlugin) Bind(wctx *WasmProcessContext) {
	importedModules := wctx.GetImportedModules()
	if _, ok := importedModules["my-own-cluster"]; !ok {
		return
	}

	if wctx.Fctx.Trace {
		fmt.Println("binding wasm api providers...")
	}

	for module, _ := range importedModules {
		apiProvider := wctx.Fctx.Orchestrator.GetAPIProvider(module)
		if apiProvider == nil {
			fmt.Printf("error : cannot bind module '%s' because api provider is not found. It might be found in stored functions though...\n", module)
			continue
		}

		apiProvider.BindToExecutionEngineContext(wctx)
	}
}
