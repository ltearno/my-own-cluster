/// reference path="./core-api-guest.d.ts"

console.log("dashboard invocation")

const moc = requireApi('core')

function invoke() {
    var status = JSON.parse(moc.getStatus());

    console.log(JSON.stringify(status))

    var resultText = "coucou"

    moc.writeExchangeBufferHeader(moc.getOutputBufferId(), "Content-Type", "text/html");
    moc.writeExchangeBuffer(moc.getOutputBufferId(), resultText);

    return 200
}