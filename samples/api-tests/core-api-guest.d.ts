// type definitions for module 'core'
//
//  you can use them by adding this at the beginning of your js file :
//  /// reference path="./core-api-guest.d.ts"
//
declare function requireApi(name: "core") : {
    getInputBufferId() : number
    getOutputBufferId() : number
    createExchangeBuffer() : number
    writeExchangeBuffer(bufferId: number, content: Uint8Array) : number
    writeExchangeBufferHeader(bufferId: number, name: string, value: string) : number
    getExchangeBufferSize(bufferId: number) : number
    // returns the written size if result_bytes was not NULL and the exchange buffer size otherwise
    readExchangeBuffer(bufferId: number) : Uint8Array
    // returns the buffer headers in JSON format
    readExchangeBufferHeaders(bufferId: number) : { [key: string]: string }
    base64Decode(encoded: string) : Uint8Array
    base64Encode(input: Uint8Array) : string
    registerBlobWithName(name: string, contentType: string, content: Uint8Array) : string
    registerBlob(contentType: string, content: Uint8Array) : string
    getBlobTechIdFromName(name: string) : string
    getBlobBytesAsString(name: string) : string
    plugFunction(method: string, path: string, name: string, startFunction: string) : number
    plugFile(method: string, path: string, name: string) : number
    getStatus() : string
    persistenceSet(key: Uint8Array, value: Uint8Array) : number
    getUrl(url: string) : Uint8Array
    persistenceGet(key: Uint8Array) : Uint8Array
    persistenceGetSubset(prefix: string) : { [key: string]: string }
    printDebug(text: string) : number
    getTime(dest: Uint8Array) : number
    freeBuffer(bufferId: number) : number
    callFunction(name: string, startFunction: string, arguments: int[], mode: string, inputExchangeBufferId: number, outputExchangeBufferId: number, posixFileName: string, posixArguments: string[]) : number
}