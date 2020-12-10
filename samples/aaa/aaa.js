/// reference path="./core-api-guest.d.ts"
/// reference path="./jwt-api-guest.d.ts"

const moc = requireApi('core')
const jwt = requireApi('jwt')

console.log("filter module")

function filter() {
    console.log("filter function")

    var headers = moc.readExchangeBufferHeaders(moc.getInputBufferId())
    var ah = headers["authorization"]

    console.log(headers["x-moc-url-path"])
    console.log(JSON.stringify(headers, null, 2))

    // get the accessed plug info (tags)
    // in cookie or header
    // find the jwt
    // verify it

    if ((typeof ah === 'string') && ah.startsWith("Bearer ")) {
        jwtToken = ah.substr(7)
        console.log("we have a jwt " + jwtToken)
        var verified = jwt.verifyJwt(jwtToken)
        console.log("verified: " + verified)
    }
    else {
        console.log("no authentication")
    }

    moc.writeExchangeBufferHeader(moc.getOutputBufferId(), "set-cookie", "Authorization=Bearer.eyJhbGciOiJSUzI1NiIsImtpZCI6Ik4wSWFidlVDWHI0VXlXZU5scGhPRlZsOFU5ZWRGN2ZLdDZvaVZXaWxWZlJsOEI3UExJc0YwdVVnV2JINHctX3BDNmdWX2k3OHl1bnNOS3o4Ml90bkRnIiwidHlwIjoiSldUIn0.eyJpc3MiOiJodHRwczovL2hvbWUubHRlY29uc3VsdGluZy5mciIsInN1YiI6IkxURSBDb25zdWx0aW5nIiwiYXVkIjoiSURQIiwiZXhwIjoxNTk1MzcyMTkyLCJpYXQiOjE1OTUzNTc3OTIsImp0aSI6IjBjOGMzMDUwLWVjMTEtNGZlNC1hZGJjLWI1ZmQ5ZDNlNDE0NCIsInV1aWQiOiJsdGVhcm5vIiwicm9sZSI6Int9Iiwicm9sZXMiOiJ7fSJ9.cq2J09_wbDNeAw2Rt5ZVEFwZbKW8psZ0j5mvCoBrrxTgbuOL1fs4_lmnZf8cz3VJoVCsK-X_NZN9njE-RGvT_P6Pi1djTTYTAxajIxSncWfob0tSUfxBRldC5-QpH8DGEGTJLqv0ZNsTjrHFissccSbECRjai5l5h6M1a_gZR6zK_MluyuYOPQ26C5ldWNnDvbshyzGMNfQSyqWgpXfF94tW0jTr-nIjlWgSPKVZF2jSiME9-WNy96OaIx3gBmxlGZDy0hA-v_zhfuMTDNN1Bwl_OUqfkl0QoFfolQeJsP173jtxVyEBGTmP_PtUz_yb_UcD25DeZrPe0zw0KlKABQ; Secure; HttpOnly; SameSite=Strict; MaxAge=10200; Path=/")
}