/**
 * This is the Core REST API my-own-cluster implementation
 * 
 * It allows to execute basic operations like :
 * 
 * - register a blob
 * - plus a blob as a function
 * - plug a blob as a file
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

    moc.writeExchangeBufferFromString(moc.getOutputBufferId(), JSON.stringify({
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

    moc.writeExchangeBufferFromString(moc.getOutputBufferId(), JSON.stringify({
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

    moc.writeExchangeBufferFromString(moc.getOutputBufferId(), JSON.stringify({
        status: true,
        tech_id: techID
    }))

    return 200
}