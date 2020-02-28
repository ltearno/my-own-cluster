function test() {
    var size = 1024
    var input = new Float32Array(size)
    for (var i = 0; i < size; i++)
        input[i] = i;
    var inputExchangeBufferId = moc.createExchangeBuffer()
    moc.writeExchangeBuffer(inputExchangeBufferId, input)

    var outputExchangeBufferId = moc.createExchangeBuffer()

    var res = moc.callFunction(
        "gpu-compute-shader",
        "",
        [],
        "direct",
        inputExchangeBufferId,
        outputExchangeBufferId,
        "",
        [""]
    )

    moc.writeExchangeBuffer(moc.getOutputBufferId(), res.output)
}