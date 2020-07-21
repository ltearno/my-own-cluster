/// reference path="./core-api-guest.d.ts"

console.log("reverseproxy loading")

const moc = requireApi('core')

function invoke() {
    try {
        var headers = moc.readExchangeBufferHeaders(moc.getInputBufferId());
        var proxyData = JSON.parse(headers["x-moc-plug-data"])

        var requestedMethod = headers["x-moc-method"] || ""
        var requestedPath = headers["x-moc-path-param-path"] || ""
        var proxyBaseUrl = proxyData.backend
        var url = proxyBaseUrl
        if (requestedPath && requestedPath.length > 0)
            url = url + "/" + requestedPath

        console.log("requestedMethod: " + requestedMethod)
        console.log("requestedPath: " + requestedPath)
        console.log("proxyBaseUrl: " + proxyBaseUrl)
        console.log("proxying to " + url)

        // proxyfies the request whether it be http, websockets, ...
        moc.betaWebProxy(JSON.stringify({
            method: requestedMethod,
            url: url,
            headers: {
                "content-type": headers["content-type"] || "application/octet-stream"
            },
            inputBufferId: moc.getInputBufferId(),
            outputBufferId: moc.getOutputBufferId()
        }))

        //var result = moc.getUrl(url)
        //moc.writeExchangeBufferHeader(moc.getOutputBufferId(), "Content-Type", "application/json")
        //moc.writeExchangeBuffer(moc.getOutputBufferId(), result)
    }
    catch (e) {
        console.log(e)
        moc.writeExchangeBufferHeader(moc.getOutputBufferId(), "content-type", "application/json")
        moc.writeExchangeBufferStatusCode(moc.getOutputBufferId(), 500)
        moc.writeExchangeBuffer(moc.getOutputBufferId(), JSON.stringify({ error: "error while proxying something" }))
    }
}

function webSocketEchoServer() {

}
