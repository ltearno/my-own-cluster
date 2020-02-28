function test() {
    console.log("test openGL")
    var size = 1024
    var input = new Float32Array(size)
    console.log(input)
    for (var i = 0; i < size; i++)
        input[i] = i

    moc.callFunction(
        "gpu-compute-shader",
        "",
        [],
        "direct",
        input,
        "",
        [""]
    )
}