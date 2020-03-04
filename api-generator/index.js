// TODO should output rust guest bindings as well

const fs = require('fs')

const log = console.log.bind(console)

function mapReturnType(tag) {
    switch (tag) {
        case "int":
            return "i"
        case "bytes":
            return "i" // returns the result buffer length if passed buffer is NULL and written size if not
        case "buffer":
            return "i" // returns the result exchange buffer id
    }

    throw `unknown return type '${tag}'`
}

function mapArgumentType(tag) {
    switch (tag) {
        case "int":
            return "i"

        case "buffer":
            return "ii"

        case "bytes":
            return "ii"

        case "string":
            return "ii"
    }

    throw `unknown return type '${tag}'`
}

function mapGoName(input = "") {
    let result = ""
    let i = 0
    while (i < input.length) {
        if (input[i] == "_") {
            i++
            result += input[i].toLocaleUpperCase()
            i++

        } else {
            let c = input[i]
            if (i == 0)
                c = c.toLocaleUpperCase()
            result += c
            i++
        }
    }
    if (result.endsWith("Id"))
        result = result.substr(0, result.length - 2) + "ID"
    return result
}

function camelCase(input = "") {
    let result = ""
    let i = 0
    while (i < input.length) {
        if (input[i] == "_") {
            i++
            result += input[i].toLocaleUpperCase()
            i++

        } else {
            result += input[i]
            i++
        }
    }
    return result
}

function mapJsName(input = "") {
    return camelCase(input)
}

function getGoParamExtractionCode(args) {
    let currentWasmParamIndex = 0
    let code = ""
    let argNames = []
    for (let i = 0; i < args.length; i++) {
        let arg = args[i]
        let argName = camelCase(arg.name)
        argNames.push(argName)
        code += `${argName} := `
        switch (arg.type) {
            case "int":
                code += `cs.GetParamInt(${currentWasmParamIndex})`
                currentWasmParamIndex++
                break
            case "buffer":
                code += `cs.GetParamByteBuffer(${currentWasmParamIndex}, ${currentWasmParamIndex + 1})`
                currentWasmParamIndex += 2
                break
            case "bytes":
                code += `cs.GetParamByteBuffer(${currentWasmParamIndex}, ${currentWasmParamIndex + 1})`
                currentWasmParamIndex += 2
                break
            case "string":
                code += `cs.GetParamString(${currentWasmParamIndex}, ${currentWasmParamIndex + 1})`
                currentWasmParamIndex += 2
                break
        }
        code += "\n"
    }
    return {
        code,
        argNames,
        currentWasmParamIndex
    }
}

function getJsParamExtractionCode(args) {
    let code = ""
    let argNames = []
    for (let i = 0; i < args.length; i++) {
        let stackPosition = -(args.length - i)
        let arg = args[i]
        let argName = camelCase(arg.name)
        argNames.push(argName)
        code += `${argName} := `
        switch (arg.type) {
            case "int":
                code += `int(c.GetNumber(${stackPosition}))`
                break
            case "buffer":
                code += `c.SafeToBytes(${stackPosition})`
                break
            case "bytes":
                code += `c.SafeToBytes(${stackPosition})`
                break
            case "string":
                code += `c.SafeToString(${stackPosition})`
                break
        }
        code += "\n"
    }
    return {
        code,
        argNames
    }
}

function getJsReturnCode(type) {
    switch (type) {
        case "int":
            return `c.PushInt(res)`

        case "buffer":
            return `dest := (*[1 << 30]byte)(c.PushBuffer(len(res), false))[:len(res):len(res)]
                copy(dest, res)`

        case "bytes":
            return `dest := (*[1 << 30]byte)(c.PushBuffer(len(res), false))[:len(res):len(res)]
                    copy(dest, res)`
    }

    throw `unknown js retrun type`
}

function getGoPreCallCode(fct, goParamExtraction) {
    switch (fct.returnType) {
        case "int":
            return ""

        case "buffer":
            return ""

        case "bytes":
            return `resultBuffer := cs.GetParamByteBuffer(${goParamExtraction.currentWasmParamIndex}, ${goParamExtraction.currentWasmParamIndex + 1})`
    }

    throw `unknown js retrun type for precall code`
}

function getGoReturnCode(fct) {
    switch (fct.returnType) {
        case "int":
            return `return uint32(res), err`

        case "buffer":
            return `
                resultBufferID := wctx.Fctx.Orchestrator.CreateExchangeBuffer()
                resultBuffer := wctx.Fctx.Orchestrator.GetExchangeBuffer(resultBufferID)
                resultBuffer.Write(res)
                return uint32(resultBufferID), nil`

        case "bytes":
            return `
                if resultBuffer != nil && len(resultBuffer)>=len(res) {
                    copy(resultBuffer, res)
                    return uint32(len(res)), nil
                } else {
                    return uint32(len(res)), nil
                }`
    }

    throw `unknown js retrun type for go return code`
}

function getWasmAdditionalPrototype(type) {
    switch (type) {
        case "bytes":
            return `ii`
    }
    return ""
}

// for each api function :
// output code for binding wasm engine to core api
// output code for binding js engine to core api
function generateWasmBindings(apiDescription, out) {
    out(`package ${apiDescription.targetGoPackage}

    import (
        "my-own-cluster/enginewasm"
    )
    \n\n`)
    out(`func ${apiDescription.bindFunctionName}WASM(wctx enginewasm.WasmProcessContext) {`)
    for (let fct of apiDescription.functions) {
        let wasmName = fct.name
        let goName = mapGoName(fct.name)

        let goParamExtraction = getGoParamExtractionCode(fct.args)
        let goCallParams = ["wctx.Fctx"]
        goCallParams = goCallParams.concat(...goParamExtraction.argNames)
        out(`
	wctx.BindAPIFunction("${apiDescription.wasmDeclaredModule}", "${wasmName}", "${mapReturnType(fct.returnType)}(${fct.args.map(arg => mapArgumentType(arg.type)).join('')}${getWasmAdditionalPrototype(fct.returnType)})", func(wctx *enginewasm.WasmProcessContext, cs *enginewasm.CallSite) (uint32, error) {
        ${goParamExtraction.code}

        ${getGoPreCallCode(fct, goParamExtraction)}

        res, err := ${goName}(${goCallParams.join(', ')})
        if err != nil {
            return uint32(0xffff), err
        }
        
        ${getGoReturnCode(fct)}
    })
    `)
    }
    out(`}`)
}

function generateJsBindings(apiDescription, out) {
    out(`package ${apiDescription.targetGoPackage}

    import (
        "my-own-cluster/enginejs"
    
        "gopkg.in/ltearno/go-duktape.v3"
    )\n\n`)
    out(`func ${apiDescription.bindFunctionName}Js(ctx enginejs.JSProcessContext) {`)
    for (let fct of apiDescription.functions) {
        let jsName = mapJsName(fct.name)
        let goName = mapGoName(fct.name)

        let jsParamExtraction = getJsParamExtractionCode(fct.args)
        let jsCallParams = ["ctx.Fctx"]
        jsCallParams = jsCallParams.concat(...jsParamExtraction.argNames)
        out(`
        ctx.Context.PushGoFunction(func(c *duktape.Context) int {
            ${jsParamExtraction.code}
            res, err := ${goName}(${jsCallParams.join(', ')})
            if err != nil {
                return 0
            }
            
            ${getJsReturnCode(fct.returnType)}
    
            return 1
        })
        ctx.Context.PutPropString(-2, "${jsName}")
        `)
    }
    out(`}`)
}

function makeOut(fileName) {
    if (fs.existsSync(fileName))
        fs.unlinkSync(fileName)
    return function (txt) {
        fs.appendFileSync(fileName, txt, { encoding: 'utf8' })
    }
}

// read api description
let apiDescription = JSON.parse(fs.readFileSync("../src/my-own-cluster/apicore/api.json"))
generateJsBindings(apiDescription, makeOut("../src/my-own-cluster/apicore/core-api-js.go"))
generateWasmBindings(apiDescription, makeOut("../src/my-own-cluster/apicore/core-api-wasm.go"))

apiDescription = JSON.parse(fs.readFileSync("../src/my-own-cluster/apigpu/api.json"))
generateJsBindings(apiDescription, makeOut("../src/my-own-cluster/apigpu/opengl-api-js.go"))
generateWasmBindings(apiDescription, makeOut("../src/my-own-cluster/apigpu/opengl-api-wasm.go"))