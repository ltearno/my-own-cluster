function test() {
    console.log("test openGL")
    var size = 1024
    var input = new Float32Array(size)
    for (var i = 0; i < size; i++)
    input[i] = i
    console.log(input)

    var res = moc.callFunction(
        "gpu-compute-shader",
        "",
        [],
        "direct",
        input,
        "",
        [""]
    )

    moc.writeExchangeBuffer(moc.getOutputBufferId(), res.output)
}