package enginewasmer

import (
	"errors"

	"github.com/ltearno/my-own-cluster/common"
	"github.com/wasmerio/wasmer-go/wasmer"
)

type WasmWasmerEngine struct {
}

func NewWasmWasmerEngine() *WasmWasmerEngine {
	return &WasmWasmerEngine{}
}

func (e *WasmWasmerEngine) PrepareContext(fctx *common.FunctionExecutionContext) (common.ExecutionEngineContext, error) {
	wctx := CreateWasmWasmerContext(fctx)
	if wctx == nil {
		return nil, errors.New("cannot create wasm context")
	}

	return wctx, nil
}

type WasmProcessContext struct {
	Fctx *common.FunctionExecutionContext

	// the engine or something to get to it
	engine *wasmer.Engine
	store  *wasmer.Store
}

func CreateWasmWasmerContext(fctx *common.FunctionExecutionContext) *WasmProcessContext {
	engine := wasmer.NewEngine()
	store := wasmer.NewStore(engine)

	return &WasmProcessContext{
		Fctx:   fctx,
		engine: engine,
		store:  store,
	}
}

func (ctx *WasmProcessContext) Run() error {
	println("we should run now")
	wasmBytes := ctx.Fctx.CodeBytes

	module, err := wasmer.NewModule(ctx.store, wasmBytes)
	if err != nil {
		println("KGHDZKJGHDZKJDGH 1")
		return err
	}

	get_input_buffer_idFunction := wasmer.NewFunction(
		ctx.store,
		wasmer.NewFunctionType(wasmer.NewValueTypes(), wasmer.NewValueTypes(wasmer.I32)),
		func(values []wasmer.Value) ([]wasmer.Value, error) {
			return []wasmer.Value{wasmer.NewI32(ctx.Fctx.InputExchangeBufferID)}, nil
		})

	importObject := wasmer.NewImportObject()
	importObject.Register("core",
		map[string]wasmer.IntoExtern{
			"get_input_buffer_id":               get_input_buffer_idFunction,
			"get_output_buffer_id":              get_input_buffer_idFunction,
			"create_exchange_buffer":            get_input_buffer_idFunction,
			"write_exchange_buffer":             get_input_buffer_idFunction,
			"write_exchange_buffer_header":      get_input_buffer_idFunction,
			"write_exchange_buffer_status_code": get_input_buffer_idFunction,
			"read_exchange_buffer":              get_input_buffer_idFunction,
			"read_exchange_buffer_headers":      get_input_buffer_idFunction,
			"base64_decode":                     get_input_buffer_idFunction,
			"base64_encode":                     get_input_buffer_idFunction,
			"register_blob_with_name":           get_input_buffer_idFunction,
			"register_blob":                     get_input_buffer_idFunction,
			"get_blob_tech_id_from_name":        get_input_buffer_idFunction,
			"get_blob_bytes_as_string":          get_input_buffer_idFunction,
			"plug_function":                     get_input_buffer_idFunction,
			"plug_file":                         get_input_buffer_idFunction,
			"unplug_path":                       get_input_buffer_idFunction,
			"get_status":                        get_input_buffer_idFunction,
			"persistence_set":                   get_input_buffer_idFunction,
			"get_url":                           get_input_buffer_idFunction,
			"persistence_get":                   get_input_buffer_idFunction,
			"persistence_get_subset":            get_input_buffer_idFunction,
			"print_debug":                       get_input_buffer_idFunction,
			"get_time":                          get_input_buffer_idFunction,
			"free_buffer":                       get_input_buffer_idFunction,
			"call_function":                     get_input_buffer_idFunction,
			"export_database":                   get_input_buffer_idFunction,
			"beta_web_proxy":                    get_input_buffer_idFunction,
			"is_trace":                          get_input_buffer_idFunction,
			"plug_filter":                       get_input_buffer_idFunction,
			"unplug_filter":                     get_input_buffer_idFunction,
		})
	instance, err := wasmer.NewInstance(module, importObject)
	if err != nil {
		println("KGHDZKJGHDZKJDGH 2 ", err.Error())
		return err
	}

	// getStatus() -> u32 in the Dashboard sample project
	aFunctionInMyWasmModule, err := instance.Exports.GetFunction("getStatus")
	if err != nil {
		println("KGHDZKJGHDZKJDGH 3")
		return err
	}

	println("call funcTION")

	aFunctionInMyWasmModule()

	return nil
}

/*
func main() {
	wasmBytes, _ := ioutil.ReadFile("simple.wasm")

	engine := wasmer.NewEngine()
	store := wasmer.NewStore(engine)

	// Compiles the module
	module, _ := wasmer.NewModule(store, wasmBytes)

	// Instantiates the module
	importObject := wasmer.NewImportObject()
	instance, _ := wasmer.NewInstance(module, importObject)

	// Gets the `sum` exported function from the WebAssembly instance.
	sum, _ := instance.Exports.GetFunction("sum")

	// Calls that exported function with Go standard values. The WebAssembly
	// types are inferred and values are casted automatically.
	result, _ := sum(5, 37)

	fmt.Println(result) // 42!
}
*/
