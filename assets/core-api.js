const log = console.log.bind(console)

function registerFunction() {
    var req = JSON.parse(moc.readExchangeBufferAsString(moc.getInputBufferId()))

    var codeBytes = moc.base64Decode(req.wasm_bytes)

    var techID = moc.registerFunction(
        req.name,
        req.type,
        codeBytes
    )

    moc.writeExchangeBufferFromString(moc.getOutputBufferId(), JSON.stringify({
        status: true,
        tech_id: techID,
        type: req.type,
        name: req.name,
        wasm_bytes_size: codeBytes.length,
    }))

    return 200
}

function plugFunction() {
    var req = JSON.parse(moc.readExchangeBufferAsString(moc.getInputBufferId()))

    var techID = moc.plugFunction(
        req.method,
        req.path,
        req.name,
        req.start_function
    )

    moc.writeExchangeBufferFromString(moc.getOutputBufferId(), JSON.stringify({
        status: true,
        tech_id: techID,
        method: req.method,
        path: req.path,
        name: req.name,
        start_function: req.start_function
    }))

    return 200
}

function plugFile() {
    var req = JSON.parse(moc.readExchangeBufferAsString(moc.getInputBufferId()))

    var fileBytes = moc.base64Decode(req.bytes)

    var techID = moc.plugFile(
        req.method,
        req.path,
        req.content_type,
        fileBytes
    )

    moc.writeExchangeBufferFromString(moc.getOutputBufferId(), JSON.stringify({
        status: true,
        tech_id: techID,
        method: req.method,
        path: req.path,
        content_type: req.content_type,
        bytes_size: fileBytes.length,
    }))

    return 200
}