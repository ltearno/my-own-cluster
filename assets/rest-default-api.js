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

const moc = require('core')

function getInputRequest() {
    var bufferId = moc.getInputBufferId()
    var bufferBytes = moc.readExchangeBuffer(bufferId)

    var reqText = new TextDecoder("utf-8").decode(bufferBytes)

    return JSON.parse(reqText)
}

function plugFunction() {
    var req = getInputRequest()

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
    var req = getInputRequest()

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
    var req = getInputRequest()

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
    var req = getInputRequest()

    var inputExchangeBufferId = moc.createExchangeBuffer()
    if (req.input)
        moc.writeExchangeBuffer(inputExchangeBufferId, moc.base64Decode(req.input))
    
    var outputExchangeBufferId = moc.createExchangeBuffer()

    var result = moc.callFunction(
        req.name,
        req.start_function || "_start",
        req.arguments || [],
        req.mode || "direct",
        inputExchangeBufferId,
        outputExchangeBufferId,
        req.posix_file_name || "",
        req.posix_arguments || [""]
    )

    // read output buffer and base 64 encode it
    result.output = moc.base64Encode(result.output)

    moc.writeExchangeBuffer(moc.getOutputBufferId(), JSON.stringify(result))

    return 200
}