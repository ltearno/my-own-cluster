
// type definitions for module 'jwt'
//
//  you can use them by adding this at the beginning of your js file :
//  /// reference path="./jwt-api-guest.d.ts"
//
//  then, simply import the module through the runtime API
//  const jwt = requireApi("jwt")
//
declare function requireApi(name: "jwt") : {
    verifyJwt(jwt: string) : string
}