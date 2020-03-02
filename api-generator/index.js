// TODO should output rust guest bindings as well

const fs = require('fs')

const log = console.log.bind(console)

// read api description
let apiDescription = JSON.parse(fs.readFileSync("my-own-cluster.api.json"))

function mapReturnType(tag) {
    switch (tag) {
        case "int":
            return "i"
        case "buffer":
            return "i" // returns the result buffer length
    }

    throw `unknown return type '${tag}'`
}

function mapArgumentType(tag) {
    switch (tag) {
        case "int":
            return "i"

        case "buffer":
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
                code += `SafeToBytes(c, ${stackPosition})`
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
    }

    throw `unknown js retrun type`
}

// for each api function :
// output code for binding wasm engine to core api
// output code for binding js engine to core api
for (let fct of apiDescription.functions) {
    let wasmName = fct.name
    let jsName = mapJsName(fct.name)
    let goName = mapGoName(fct.name)

    let goParamExtraction = getGoParamExtractionCode(fct.args)
    let goCallParams = ["wctx.Fctx"]
    goCallParams = goCallParams.concat(...goParamExtraction.argNames)
    log(`
    // wasm params : ${fct.args.map(arg => arg.name).join(' ')} ${fct.returnType=='buffer'?`result_buffer_addr result_buffer_length`:''}
	wctx.BindAPIFunction("${apiDescription.wasmDeclaredModule}", "${wasmName}", "${mapReturnType(fct.returnType)}(${fct.args.map(arg => mapArgumentType(arg.type)).join('')}${fct.returnType=='buffer'?'ii':''})", func(wctx *WasmProcessContext, cs *CallSite) (uint32, error) {
        ${goParamExtraction.code}

        ${fct.returnType!='buffer'?'':`resultBuffer := cs.GetParamByteBuffer(${goParamExtraction.currentWasmParamIndex}, ${goParamExtraction.currentWasmParamIndex + 1})`}

        res, err := ${apiDescription.targetGoPackage}.${goName}(${goCallParams.join(', ')})
        
        ${fct.returnType!='buffer'?`return uint32(res), err`:`if resultBuffer != nil && len(resultBuffer)==len(res) {\n                copy(resultBuffer, res)\n        }\n        return uint32(len(res)), err`}
    })
    `)

    let jsParamExtraction = getJsParamExtractionCode(fct.args)
    let jsCallParams = ["fctx"]
    jsCallParams = jsCallParams.concat(...jsParamExtraction.argNames)
    log(`
    ctx.PushGoFunction(func(c *duktape.Context) int {
        ${jsParamExtraction.code}
        res, err := ${apiDescription.targetGoPackage}.${goName}(${jsCallParams.join(', ')})
        if err != nil {
            return 0
        }
        
        ${getJsReturnCode(fct.returnType)}

        return 1
    })
    ctx.PutPropString(-2, "${jsName}")
    `)
}