// TODO should output rust guest bindings as well

const fs = require('fs')

function mapReturnType(tag) {
    switch (tag) {
        case "int":
            return "i"
        case "bytes":
            return "i" // returns the result buffer length if passed buffer is NULL and written size if not
        case "string":
            return "i" // returns the result buffer length if passed buffer is NULL and written size if not
        case "buffer":
            return "i" // returns the result exchange buffer id
        case "map[string]string":
            return "i" // exchange buffer id
    }

    throw `unknown return type '${tag}'`
}

function mapArgumentType(tag) {
    switch (tag) {
        case "int":
            return "i"

        case "bytes":
            return "ii"

        case "string":
            return "ii"

        case "[]int":
            return "ii"

        case "[]string":
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

            case "bytes":
                code += `cs.GetParamByteBuffer(${currentWasmParamIndex}, ${currentWasmParamIndex + 1})`
                currentWasmParamIndex += 2
                break

            case "string":
                code += `cs.GetParamString(${currentWasmParamIndex}, ${currentWasmParamIndex + 1})`
                currentWasmParamIndex += 2
                break

            case "[]int":
                code += `[]int{} // TODO : To be implemented !!!`
                currentWasmParamIndex += 2
                break

            case "[]string":
                code += `[]string{} // TODO : To be implemented !!!`
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

            case "bytes":
                code += `c.SafeToBytes(${stackPosition})`
                break

            case "string":
                code += `c.SafeToString(${stackPosition})`
                break

            case "[]int":
                code += `[]int{}
                // #define DUK_ENUM_ARRAY_INDICES_ONLY       (1U << 5)    // only enumerate array indices
                c.Enum(${stackPosition}, (1 << 5))
                for c.Next(-1, true) {
                    ${argName} = append(${argName}, c.GetInt(-1))
                    c.Pop()
                    c.Pop()
                }
                c.Pop()`
                break

            case "[]string":
                code += `[]string{}
                // #define DUK_ENUM_ARRAY_INDICES_ONLY       (1U << 5)    // only enumerate array indices
                c.Enum(${stackPosition}, (1 << 5))
                for c.Next(-1, true) {
                    ${argName} = append(${argName}, c.SafeToString(-1))
                    c.Pop()
                    c.Pop()
                }
                c.Pop()`
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

        case "string":
            return `c.PushString(res)`

        case "bytes":
            return `dest := (*[1 << 30]byte)(c.PushBuffer(len(res), false))[:len(res):len(res)]
                    copy(dest, res)`

        case "map[string]string":
            return `
            c.PushObject()
            for k, v := range res {
                c.PushString(v)
                c.PutPropString(-2, k)
            }`
    }

    throw `unknown js retrun type`
}

function getGoPreCallCode(fct, goParamExtraction) {
    switch (fct.returnType) {
        case "int":
            return ""

        case "buffer":
            return ""

        case "string":
            return ""

        case "bytes":
            return `resultBuffer := cs.GetParamByteBuffer(${goParamExtraction.currentWasmParamIndex}, ${goParamExtraction.currentWasmParamIndex + 1})`

        case "map[string]string":
            return ""
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

        case "string":
            return `
                resultBufferID := wctx.Fctx.Orchestrator.CreateExchangeBuffer()
                resultBuffer := wctx.Fctx.Orchestrator.GetExchangeBuffer(resultBufferID)
                resultBuffer.Write([]byte(res))
                return uint32(resultBufferID), nil`

        case "bytes":
            return `
                if resultBuffer != nil && len(resultBuffer)>=len(res) {
                    copy(resultBuffer, res)
                    return uint32(len(res)), nil
                } else {
                    return uint32(len(res)), nil
                }`

        case "map[string]string":
            return `
                var b bytes.Buffer
                bs := make([]byte, 4)

                // pair count
                binary.LittleEndian.PutUint32(bs, uint32(len(res)))
                b.Write(bs)

                for k, v := range res {
                    // write key
                    binary.LittleEndian.PutUint32(bs, uint32(len([]byte(k))))
                    b.Write(bs)
                    b.Write([]byte(k))

                    // write value
                    binary.LittleEndian.PutUint32(bs, uint32(len([]byte(v))))
                    b.Write(bs)
                    b.Write([]byte(v))
                }

                resultBufferID := wctx.Fctx.Orchestrator.CreateExchangeBuffer()
                resultBuffer := wctx.Fctx.Orchestrator.GetExchangeBuffer(resultBufferID)
                resultBuffer.Write(b.Bytes())

                return uint32(resultBufferID), nil`
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
    let needBytesPackage = Object.values(apiDescription.functions).some(fct => fct.returnType == "map[string]string")
    out(`package api${apiDescription.targetGoPackage}

    import (
        "my-own-cluster/enginewasm"

        ${needBytesPackage ? '"bytes"' : ''}
        ${needBytesPackage ? '"encoding/binary"' : ''}
    )
    \n\n`)
    out(`func ${apiDescription.bindFunctionName}WASM(wctx enginewasm.WasmProcessContext) {`)
    for (let fctName in apiDescription.functions) {
        let fct = apiDescription.functions[fctName]
        let wasmName = fctName
        let goName = mapGoName(fctName)

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
    out(`package api${apiDescription.targetGoPackage}

    import (
        "my-own-cluster/enginejs"
    
        "gopkg.in/ltearno/go-duktape.v3"
    )\n\n`)
    out(`func ${apiDescription.bindFunctionName}Js(ctx enginejs.JSProcessContext) {`)
    for (let fctName in apiDescription.functions) {
        let fct = apiDescription.functions[fctName]

        let jsName = mapJsName(fctName)
        let goName = mapGoName(fctName)

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

function getTypescriptArgument(arg) {
    return `${mapJsName(arg.name)}: ${mapTypescriptType(arg.type)}`
}

function mapTypescriptType(tag) {
    switch (tag) {
        case "int":
            return "number"

        case "bytes":
            return "Uint8Array"

        case "buffer":
            return "Uint8Array"

        case "string":
            return "string"

        case "[]int":
            return "int[]"

        case "[]string":
            return "string[]"

        case "map[string]string":
            return "{ [key: string]: string }"
    }

    throw `unknown return type '${tag}'`
}

function generateGuestTypescriptBindings(apiDescription, out) {
    out(`
// type definitions for module '${apiDescription.targetGoPackage}'
//
//  you can use them by adding this at the beginning of your js file :
//  /// reference path="./${apiDescription.targetGoPackage}-api-guest.d.ts"
//
declare function requireApi(name: "${apiDescription.targetGoPackage}") : {
`)

    for (let fctName in apiDescription.functions) {
        let fct = apiDescription.functions[fctName]

        let jsName = mapJsName(fctName)
        if (fct.comment)
            out(`    // ${fct.comment}\n`)
        out(`    ${jsName}(${fct.args.map(getTypescriptArgument).join(', ')}) : ${mapTypescriptType(fct.returnType)}\n`)
    }

    out(`}`)
}

function getCReturnType(type) {
    switch (type) {
        case "int":
            return "uint32_t"
        case "bytes":
            return "uint32_t" // returns the result buffer length if passed buffer is NULL and written size if not
        case "string":
            return "uint32_t" // returns the result exhcnage buffer id
        case "buffer":
            return "uint32_t" // returns the result exchange buffer id
        case "map[string]string":
            return "uint32_t" // return exchange buffer id
    }

    throw `unknown return type '${tag}'`
}

function getCArgument(arg) {
    switch (arg.type) {
        case "int":
            return `int ${arg.name}`

        case "bytes":
            return `const void *${arg.name}_bytes, int ${arg.name}_length`

        case "string":
            return `const char *${arg.name}_string, int ${arg.name}_length`

        case "[]int":
            return `const void *${arg.name}_int_array, int ${arg.name}_length`

        case "[]string":
            return `const void *${arg.name}_string_array, int ${arg.name}_length`
    }

    throw `unknown return type '${tag}'`
}

function generateGuestCBindings(apiDescription, out) {
    out(`
#ifndef ${apiDescription.targetGoPackage}_api_h
#define ${apiDescription.targetGoPackage}_api_h

#include <stdint.h>

#define WASM_EXPORT                   extern __attribute__((used)) __attribute__((visibility ("default")))
#define WASM_EXPORT_AS(NAME)          WASM_EXPORT __attribute__((export_name(NAME)))
#define WASM_IMPORT(MODULE,NAME)      __attribute__((import_module(MODULE))) __attribute__((import_name(NAME)))
#define WASM_CONSTRUCTOR              __attribute__((constructor))

`)


    for (let fctName in apiDescription.functions) {
        let fct = apiDescription.functions[fctName]

        let args = fct.args.map(getCArgument)
        if (fct.returnType == "bytes")
            args.push(`void *result_bytes, int result_length`)
        if (fct.comment)
            out(`// ${fct.comment}\n`)
        out(`WASM_IMPORT("${apiDescription.wasmDeclaredModule}", "${fctName}") ${getCReturnType(fct.returnType)} ${fctName}(${args.join(', ')});\n`)
    }

    out(`
#endif
    `)
}

function generateGuestCSymsBindings(apiDescription, out) {
    for (let fctName in apiDescription.functions) {
        out(`${fctName}\n`)
    }
}

function makeOut(fileName) {
    if (fs.existsSync(fileName))
        fs.unlinkSync(fileName)
    return function (txt) {
        fs.appendFileSync(fileName, txt, { encoding: 'utf8' })
    }
}

// generate files
for (let apiName of ["core", "gpu"]) {
    let apiDescription = JSON.parse(fs.readFileSync(`../src/my-own-cluster/api${apiName}/api.json`))

    generateJsBindings(apiDescription, makeOut(`../src/my-own-cluster/api${apiName}/${apiName}-api-js.go`))

    generateWasmBindings(apiDescription, makeOut(`../src/my-own-cluster/api${apiName}/${apiName}-api-wasm.go`))

    // assets
    generateGuestCBindings(apiDescription, makeOut(`../assets/${apiName}-api-guest.h`))
    generateGuestCSymsBindings(apiDescription, makeOut(`../assets/${apiName}-api-guest.syms`))
    generateGuestTypescriptBindings(apiDescription, makeOut(`../assets/${apiName}-api-guest.d.ts`))
}