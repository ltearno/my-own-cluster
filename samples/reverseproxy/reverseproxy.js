/// reference path="./core-api-guest.d.ts"

console.log("reverseproxy loading")

const moc = requireApi('core')

function invoke() {
    var headers = moc.readExchangeBufferHeaders(moc.getInputBufferId());
    if (!("x-moc-plug-data" in headers)) {
        moc.writeExchangeBufferHeader(moc.getOutputBufferId(), "Content-Type", "application/json")
        moc.writeExchangeBuffer(moc.getOutputBufferId(), JSON.stringify({ error: "no proxying specs, you forgot plaing the proxy base url in the plug data, aborting..." }))
        return 404
    }

    var requestedPath = headers["x-moc-path-param-path"] || ""
    var proxyBaseUrl = headers["x-moc-plug-data"] || "nil://"
    var url = proxyBaseUrl + "/" + requestedPath

    console.log("proxyBaseUrl: " + proxyBaseUrl)
    console.log("requestedPath: " + requestedPath)
    console.log("proxying to " + url)

    var result = moc.getUrl(url)
    moc.writeExchangeBufferHeader(moc.getOutputBufferId(), "Content-Type", "application/json")
    moc.writeExchangeBuffer(moc.getOutputBufferId(), result)
    return 200
}