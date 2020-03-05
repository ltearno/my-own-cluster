
// type definitions for module 'gpu'
//
//  you can use them by adding this at the beginning of your js file :
//  /// reference path="./gpu-api-guest.d.ts"
//
declare function requireApi(name: "gpu") : {
    computeShader(specification: string) : number
    createImageFromRgbaFloatPixels(width: number, height: number, pixelsExchangeBufferId: number, pngExchangeBufferId: number) : number
    createImageFromRFloatPixels(width: number, height: number, pixelsExchangeBufferId: number, pngExchangeBufferId: number) : number
}