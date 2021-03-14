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
            return `if res == nil {
                    return 0
                }
                dest := (*[1 << 30]byte)(c.PushBuffer(len(res), false))[:len(res):len(res)]
                copy(dest, res)`

        case "string":
            return `c.PushString(res)`

        case "bytes":
            return `if res == nil {
                    return 0
                }
                dest := (*[1 << 30]byte)(c.PushBuffer(len(res), false))[:len(res):len(res)]
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
                binary.LittleEndian.PutUint32(bs, uint32(2*len(res)))
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
    out(`package api${apiDescription.moduleName}

    import (
        "github.com/ltearno/my-own-cluster/enginewasm"

        ${needBytesPackage ? '"bytes"' : ''}
        ${needBytesPackage ? '"encoding/binary"' : ''}
    )
    \n\n`)
    out(`func ${apiDescription.bindFunctionName}WASM(wctx enginewasm.WasmProcessContext, cookie interface{}) {`)
    for (let fctName in apiDescription.functions) {
        let fct = apiDescription.functions[fctName]
        let wasmName = fctName
        let goName = mapGoName(fctName)

        let goParamExtraction = getGoParamExtractionCode(fct.args)
        let goCallParams = ["wctx.Fctx", "cookie"]
        goCallParams = goCallParams.concat(...goParamExtraction.argNames)
        out(`
	wctx.BindAPIFunction("${apiDescription.moduleName}", "${wasmName}", "${mapReturnType(fct.returnType)}(${fct.args.map(arg => mapArgumentType(arg.type)).join('')}${getWasmAdditionalPrototype(fct.returnType)})", func(wctx *enginewasm.WasmProcessContext, cs *enginewasm.CallSite) (uint32, error) {
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
    out(`package api${apiDescription.moduleName}

    import (
        "github.com/ltearno/my-own-cluster/enginejs"
    
        "gopkg.in/ltearno/go-duktape.v3"
    )\n\n`)
    out(`func ${apiDescription.bindFunctionName}Js(ctx enginejs.JSProcessContext, cookie interface{}) {`)
    for (let fctName in apiDescription.functions) {
        let fct = apiDescription.functions[fctName]

        let jsName = mapJsName(fctName)
        let goName = mapGoName(fctName)

        let jsParamExtraction = getJsParamExtractionCode(fct.args)
        let jsCallParams = ["ctx.Fctx", "cookie"]
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
// type definitions for module '${apiDescription.moduleName}'
//
//  you can use them by adding this at the beginning of your js file :
//  /// reference path="./${apiDescription.moduleName}-api-guest.d.ts"
//
//  then, simply import the module through the runtime API
//  const ${apiDescription.moduleName} = requireApi("${apiDescription.moduleName}")
//
declare function requireApi(name: "${apiDescription.moduleName}") : {
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
#ifndef ${apiDescription.moduleName}_api_h
#define ${apiDescription.moduleName}_api_h

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
        out(`WASM_IMPORT("${apiDescription.moduleName}", "${fctName}") ${getCReturnType(fct.returnType)} ${fctName}(${args.join(', ')});\n`)
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

function getRustArgument(arg) {
    switch (arg.type) {
        case "int":
            return `${arg.name}:u32`

        case "bytes":
            return `${arg.name}_bytes: *const u8, ${arg.name}_length: u32`

        case "string":
            return `${arg.name}_string: *const u8, ${arg.name}_length: u32`

        case "[]int":
            return `${arg.name}_int_array: *const u32, ${arg.name}_length: u32`

        case "[]string":
            return `${arg.name}_string_array: *const u8, ${arg.name}_length: u32`
    }

    throw `unknown return type '${tag}'`
}




function getRustGentleArgument(arg) {
    switch (arg.type) {
        case "int":
            return `${arg.name}:u32`

        case "bytes":
            return `${arg.name}: &[u8]`

        case "string":
            return `${arg.name}: &str`

        case "[]int":
            return `${arg.name}: &[u32]`

        case "[]string":
            return `${arg.name}: &[&str]`
    }

    throw `unknown return type '${tag}'`
}

function getRustGentleArgumentTransform(arg) {
    switch (arg.type) {
        case "int":
            return `${arg.name}`

        case "bytes":
            return `${arg.name}.as_ptr(), ${arg.name}.len() as u32`

        case "string":
            return `${arg.name}.as_bytes().as_ptr(), ${arg.name}.as_bytes().len() as u32`

        case "[]int":
            return `${arg.name}.as_ptr(), ${arg.name}.len() as u32`

        case "[]string":
            return `std::ptr::null(), 0`
    }

    throw `unknown return type '${tag}'`
}


function getRustGentleReturnType(type) {
    switch (type) {
        case "int":
            return `u32`

        case "bytes":
            return `Result<Vec<u8>, u32>`

        case "buffer":
            return `Result<Vec<u8>, u32>`

        case "string":
            return `Result<String, u32>`

        case "map[string]string":
            return `Result<HashMap<String,String>, u32>`
    }

    throw `unknown return type '${type}'`
}

function getRustGentleCall(fctName, fct) {
    switch (fct.returnType) {
        case "bytes":
            return `let result_size = unsafe { raw::${fctName}(${fct.args.map(getRustGentleArgumentTransform).join(', ')}, std::ptr::null_mut(), 0) };
    let mut result = Vec::with_capacity(result_size as usize);
    result.resize(result_size as usize, 0);
    unsafe { raw::${fctName}(${fct.args.map(getRustGentleArgumentTransform).join(', ')}, result.as_mut_ptr(), result_size) };
    Ok(result)`

        case "buffer":
            return `let result_buffer_id = unsafe { raw::${fctName}(${fct.args.map(getRustGentleArgumentTransform).join(', ')}) };
    if result_buffer_id == 0xffff {
        Err(1)
    }
    else {
        let result = read_exchange_buffer(result_buffer_id);
        match result {
            Ok(result) => {
                //free_buffer(result_buffer_id);
                Ok(result)
            },
            Err(err) => {
                Err(2)
            },
        }
    }`

        case "string":
            return `let result_buffer_id = unsafe { raw::${fctName}(${fct.args.map(getRustGentleArgumentTransform).join(', ')}) };
    if result_buffer_id == 0xffff {
        Err(3)
    }
    else {
        let result_buffer = read_exchange_buffer(result_buffer_id);
        match result_buffer {
            Ok(result_buffer) => {
                let result = String::from_utf8(result_buffer).unwrap();
                //free_buffer(result_buffer_id);
                Ok(result)
            },
            Err(err) => {
                Err(4)
            },
        }
    }`

        case "map[string]string":
            return `let result_buffer_id = unsafe { raw::${fctName}(${fct.args.map(getRustGentleArgumentTransform).join(', ')}) };
    if result_buffer_id == 0xffff {
        Err(5)
    }
    else {
        let result_buffer = read_exchange_buffer(result_buffer_id);
        match result_buffer {
            Ok(result_buffer) => {
                let mut result = HashMap::new();
                let buffer = result_buffer.as_slice();
                free_buffer(result_buffer_id);

                let mut rdr = Cursor::new(buffer);
                let mut current_key : String = "".to_string();

                let mut nb_buffers = rdr.read_u32::<LittleEndian>().unwrap();
                while nb_buffers > 0 {
                    let buffer_size = rdr.read_u32::<LittleEndian>().unwrap();
                    
                    let mut buffer = Vec::with_capacity(buffer_size as usize);
                    unsafe { buffer.set_len(buffer_size as usize); }
                    rdr.read(&mut buffer);

                    let s = String::from_utf8(buffer.to_vec()).unwrap();
                    
                    if nb_buffers % 2 == 0 {
                        current_key = s;
                    } else {
                        result.insert(current_key.clone(), s);
                    }

                    nb_buffers = nb_buffers - 1;        
                }

                Ok(result)
            },
            Err(err) => {
                Err(6)
            },
        }
    }`

        case "int":
            return `unsafe { raw::${fctName}(${fct.args.map(getRustGentleArgumentTransform).join(', ')}) }`
    }

    throw `unknown gentle call '${fct.returnType}'`
}

function generateRustGentleFunction(fctName, fct, out) {
    out(`pub fn ${fctName}(${fct.args.map(getRustGentleArgument).join(', ')}) -> ${getRustGentleReturnType(fct.returnType)} {
    ${getRustGentleCall(fctName, fct)}
}

`)
}

function generateGuestRustBindings(apiDescription, out) {
    out(`
use std::collections::HashMap;
use std::io::{Cursor, Read};
use byteorder::{LittleEndian, ReadBytesExt};

/**
 * '${apiDescription.moduleName}' guest API bindings for rust
 * 
 * it can be called from rust code compiled into wasm and executed by my-own-cluster
 * 
 * you can use it by adding that in your rust source files at the beginning :
 * 
 *  mod ${apiDescription.moduleName}_api_guest;
 *  use ${apiDescription.moduleName}_api_guest::*;
 */

// import the '${apiDescription.moduleName}' module
pub mod raw {
    #[link(wasm_import_module = "${apiDescription.moduleName}")]
    extern {
`)

    for (let fctName in apiDescription.functions) {
        let fct = apiDescription.functions[fctName]

        let args = fct.args.map(getRustArgument)
        if (fct.returnType == "bytes")
            args.push(`result_bytes: *mut u8, result_length: u32`)
        if (fct.comment)
            out(`        // ${fct.comment}\n`)
        out(`        pub fn ${fctName}(${args.join(', ')}) -> u32;\n`)
    }

    out(`
    }
}

`)

    for (let fctName in apiDescription.functions) {
        let fct = apiDescription.functions[fctName]

        generateRustGentleFunction(fctName, fct, out)
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
for (let apiName of ["core", "gpu", "jwt"]) {
    let apiDescription = JSON.parse(fs.readFileSync(`../api${apiName}/api.json`))

    // bindings for host
    generateJsBindings(apiDescription, makeOut(`../api${apiName}/${apiName}-api-js.go`))
    generateWasmBindings(apiDescription, makeOut(`../api${apiName}/${apiName}-api-wasm.go`))

    // bindings for guest
    generateGuestCBindings(apiDescription, makeOut(`../assets/${apiName}-api-guest.h`))
    generateGuestCSymsBindings(apiDescription, makeOut(`../assets/${apiName}-api-guest.syms`))
    generateGuestTypescriptBindings(apiDescription, makeOut(`../assets/${apiName}-api-guest.d.ts`))
    generateGuestRustBindings(apiDescription, makeOut(`../assets/${apiName}_api_guest.rs`))
}