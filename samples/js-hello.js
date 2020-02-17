console.log("js initialisation")

function test(a) {
    //let t = "e"//JSON.stringify(JSON.parse('{"toto":"titi"}'))
    var r = JSON.stringify(JSON.parse('{"toto":"titi"}'))
    console.log("INSIDE MY YEAH FUNCTION with arg: " + a + " " + r)
}

function registerFunction() {
    var body = JSON.parse(readExchangeBufferAsString(getInputBufferId()))
    console.log("received " + JSON.stringify(body))
}