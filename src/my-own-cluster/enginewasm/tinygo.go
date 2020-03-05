package enginewasm

import (
	"fmt"
	"unsafe"

	"github.com/ltearno/go-wasm3"
)

func BindTinyGoRuntimeAPI(wctx *WasmProcessContext) {
	importedModules := wctx.GetImportedModules()
	if _, ok := importedModules["env"]; !ok {
		return
	}

	if wctx.Fctx.Trace {
		fmt.Println("binding TinyGo 0.11.0 API...")
	}

	wctx.Runtime.AttachFunction("env", "io_get_stdout", "i()", func(runtime wasm3.RuntimeT, sp unsafe.Pointer, mem unsafe.Pointer) int {
		if wctx.Fctx.Trace {
			fmt.Printf("TinyGo talks to us !!!!\n")
		}

		*(*uint32)(sp) = 1

		return 0
	})

	wctx.BindAPIFunction("env", "resource_write", "i(iii)", func(wctx *WasmProcessContext, cs *CallSite) (uint32, error) {
		fd := cs.GetParamUINT32(0)
		buffer := cs.GetParamByteBuffer(1, 2)

		if wctx.Fctx.Trace {
			fmt.Printf("TinyGo on fd %d says '%s'\n", fd, string(buffer))
		}

		return uint32(len(buffer)), nil
	})

	wctx.BindNotYetImplementedFunction("env", "runtime.ticks", "i()")
	wctx.BindNotYetImplementedFunction("env", "runtime.sleepTicks", "i(i)")
	wctx.BindNotYetImplementedFunction("env", "syscall/js.valueLength", "i(*ii)")
	wctx.BindNotYetImplementedFunction("env", "syscall/js.valueCall", "i(i*iiiiiii)")
	wctx.BindNotYetImplementedFunction("env", "syscall/js.valueIndex", "i()")
	wctx.BindNotYetImplementedFunction("env", "syscall/js.valueGet", "i()")
	wctx.BindNotYetImplementedFunction("env", "syscall/js.valueNew", "i()")
	wctx.BindNotYetImplementedFunction("env", "syscall/js.valueSet", "i()")
	wctx.BindNotYetImplementedFunction("env", "syscall/js.valueSetIndex", "i()")
	wctx.BindNotYetImplementedFunction("env", "syscall/js.stringVal", "i()")
	wctx.BindNotYetImplementedFunction("env", "syscall/js.valuePrepareString", "i()")
	wctx.BindNotYetImplementedFunction("env", "syscall/js.valueLoadString", "i()")
}
