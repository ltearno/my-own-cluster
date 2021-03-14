package enginejs

import (
	"fmt"

	"github.com/ltearno/my-own-cluster/common"

	"gopkg.in/ltearno/go-duktape.v3"
)

type JSProcessContext struct {
	Fctx *common.FunctionExecutionContext

	Context *duktape.Context
}

type JavascriptDuktapeEngine struct {
}

func NewJavascriptDuktapeEngine() *JavascriptDuktapeEngine {
	return &JavascriptDuktapeEngine{}
}

func (e *JavascriptDuktapeEngine) PrepareContext(fctx *common.FunctionExecutionContext) (common.ExecutionEngineContext, error) {
	ctx := &JSProcessContext{
		Fctx:    fctx,
		Context: duktape.New(),
	}

	ctx.Context.PushGlobalObject()

	ctx.Context.PushGoFunction(func(c *duktape.Context) int {
		module := c.SafeToString(-1)
		apiProvider := fctx.Orchestrator.GetAPIProvider(module)
		if apiProvider == nil {
			fmt.Printf("js program requires unknown module '%s'\n", module)
			return 0
		}

		ctx.Context.PushObject()

		apiProvider.BindToExecutionEngineContext(ctx)

		return 1
	})
	ctx.Context.PutPropString(-2, "requireApi")

	ctx.Context.Pop()

	return ctx, nil
}

func (jsctx *JSProcessContext) Run() error {
	if jsctx.Fctx.Trace {
		fmt.Printf("- start run js function '%s'\n", jsctx.Fctx.StartFunction)
	}

	jsctx.Context.PushString(string(jsctx.Fctx.CodeBytes))
	err := jsctx.Context.Peval()
	if err != nil {
		return fmt.Errorf("cannot eval script, probably syntax error (%v)", err)
	}

	if jsctx.Fctx.Trace {
		fmt.Println("- script eval-ed")
	}

	jsctx.Context.Pop()
	jsctx.Context.PushGlobalObject()
	ok := jsctx.Context.GetPropString(-1, jsctx.Fctx.StartFunction)
	if !ok {
		return fmt.Errorf("cannot find start function %s", jsctx.Fctx.StartFunction)
	}

	if jsctx.Fctx.Trace {
		fmt.Printf("- function '%s' found\n", jsctx.Fctx.StartFunction)
	}

	res := jsctx.Context.Pcall(0)
	if res != 0 {
		return fmt.Errorf("error while executing js function '%s'", jsctx.Fctx.StartFunction)
	}

	if jsctx.Fctx.Trace {
		fmt.Printf("- function executed\n")
	}

	jsctx.Fctx.Result = jsctx.Context.GetInt(-1)

	jsctx.Context.DestroyHeap()

	if jsctx.Fctx.Trace {
		fmt.Printf("- heap destroyed, result:%d\n", jsctx.Fctx.Result)
	}

	return nil
}
