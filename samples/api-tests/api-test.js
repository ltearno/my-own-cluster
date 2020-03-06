/// reference path="./core-api-guest.d.ts"

const log = console.log.bind(console)

log("tests api from my-own-cluster")

const moc = requireApi('core')

function run() {
    var status = {}

    moc.printDebug("you should see this message in server console")

    status.mocStatus = JSON.parse(moc.getStatus());
    log("fetching status")

    log("test 'base64'")
    const phrase = "encoding decoding works !"
    var phraseTest = new TextDecoder("utf-8").decode(moc.base64Decode(moc.base64Encode(moc.base64Decode(moc.base64Encode(phrase)))))
    status.testBase64 = phrase == phraseTest ? "ok" : ("KO " + phraseTest)

    status.inputBufferId = moc.getInputBufferId()
    status.outputBufferId = moc.getOutputBufferId()

    var bid = moc.createExchangeBuffer()
    var b = new Uint8Array(10)
    for (var i = 0; i < 10; i++) b[i] = i
    moc.writeExchangeBuffer(bid, b)
    bout = moc.readExchangeBuffer(bid)
    var ok = true
    if (!bout) ok = false;
    else {
        for (var i = 0; i < 10; i++) {
            if (bout[i] != b[i]) {
                ok = false
                break
            }
        }
    }
    status.testReadWriteBuffer = ok ? "ok" : "KO"
    status.testReadBufferSize = moc.getExchangeBufferSize(bid) == 10 ? "ok" : "KO"
    moc.freeBuffer(bid)
    bid = -1

    moc.writeExchangeBufferHeader(bid, "titi", "tutu")
    var headers = moc.readExchangeBufferHeaders(bid)
    status.testReadWriteBufferHeaders = headers["titi"] == "tutu" ? "ok" : "KO"

    moc.writeExchangeBufferHeader(moc.getOutputBufferId(), "Content-Type", "application/json");
    moc.writeExchangeBuffer(moc.getOutputBufferId(), JSON.stringify(status, null, 2));

    return 200
}