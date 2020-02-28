/**
 * This is the Core REST API my-own-cluster implementation
 * 
 * It allows to execute basic operations like :
 * 
 * - register a blob
 * - plus a blob as a function
 * - plug a blob as a file
 * - call a function
 * 
 * It is mainly used by the CLI program.
 * It is normally bound to the cluster web endpoint during startup (in main.go)
 */

const log = console.log.bind(console)

function plugFunction() {
    var req = JSON.parse(moc.readExchangeBufferAsString(moc.getInputBufferId()))

    moc.plugFunction(
        req.method,
        req.path,
        req.name,
        req.start_function
    )

    moc.writeExchangeBuffer(moc.getOutputBufferId(), JSON.stringify({
        status: true,
    }))

    return 200
}

function plugFile() {
    var req = JSON.parse(moc.readExchangeBufferAsString(moc.getInputBufferId()))

    moc.plugFile(
        req.method,
        req.path,
        req.name
    )

    moc.writeExchangeBuffer(moc.getOutputBufferId(), JSON.stringify({
        status: true,
    }))

    return 200
}

function registerBlob() {
    var req = JSON.parse(moc.readExchangeBufferAsString(moc.getInputBufferId()))

    var bytes = moc.base64Decode(req.bytes)

    var techID;

    if (req.name) {
        techID = moc.registerBlobWithName(
            req.name,
            req.content_type,
            bytes
        );
    } else {
        techID = moc.registerBlob(
            req.content_type,
            bytes
        );
    }

    moc.writeExchangeBuffer(moc.getOutputBufferId(), JSON.stringify({
        status: true,
        tech_id: techID
    }))

    return 200
}

function callFunction() {
    var req = JSON.parse(moc.readExchangeBufferAsString(moc.getInputBufferId()))

    var input = req.input ? moc.base64Decode(req.input) : null

    var result = moc.callFunction(
        req.name,
        req.start_function || "_start",
        req.arguments || [],
        req.mode || "direct",
        input,
        req.posix_file_name || "",
        req.posix_arguments || [""]
    )

    // read output buffer and base 64 encode it
    result.output = moc.base64Encode(result.output)

    moc.writeExchangeBuffer(moc.getOutputBufferId(), JSON.stringify(result))

    return 200
}