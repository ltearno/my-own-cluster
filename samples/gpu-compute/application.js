function test() {
    var dataSize = 10;
    var values = new Float32Array(dataSize);
    for (var i = 0; i < dataSize; i++) {
        values[i] = i;
    }

    var bufferId = moc.createExchangeBuffer();
    var written = moc.writeExchangeBuffer(bufferId, values);
    console.log("written " + written + " values")

    var bytes = moc.readExchangeBuffer(bufferId);
    var outValues = new Float32Array(bytes.buffer);
    for (var i = 0; i < dataSize; i++)
        console.log(outValues[i] + " / " + values[i])
}

function launchMandelbrotShader() {
    console.log("init data")
    var dataSize = 1024;
    var textureWidth = dataSize;
    var textureHeight = dataSize;
    var in1 = new Float32Array(dataSize);
    var in2 = new Float32Array(dataSize);
    var out = new Float32Array(dataSize);
    for (var i = 0; i < dataSize; i++) {
        in1[i] = i;
        in2[i] = i;
    }
    var params = new Float32Array(2)
    params[0] = 1;
    params[1] = textureWidth;
    var texture = new Float32Array(4 * textureWidth * textureHeight);

    console.log("create buffers")
    var textureBufferId = moc.createExchangeBuffer();
    // we write in the buffer because opengl api will read it to fill the texture (no option yet)
    moc.writeExchangeBuffer(textureBufferId, texture);

    var in1BufferId = moc.createExchangeBuffer();
    moc.writeExchangeBuffer(in1BufferId, in1);

    var in2BufferId = moc.createExchangeBuffer();
    moc.writeExchangeBuffer(in2BufferId, in2);

    var outBufferId = moc.createExchangeBuffer();
    moc.writeExchangeBuffer(outBufferId, out);

    var paramsBufferId = moc.createExchangeBuffer();
    moc.writeExchangeBuffer(paramsBufferId, params);

    console.log("compute shader...")
    var res = moc.computeShader(
        JSON.stringify({
            shader_name: "gpu-compute-shader",
            dispatch_size: [dataSize, dataSize, 1],
            bindings: [
                {
                    target: "TEXTURE_2D_RGBA_FLOAT",
                    exchange_buffer_id: textureBufferId,
                    is_in: false,
                    is_out: true,
                    width: textureWidth,
                    height: textureHeight
                },
                {
                    target: "STORAGE",
                    exchange_buffer_id: in1BufferId,
                    is_in: true,
                    is_out: false,
                    width: -1,
                    height: -1
                },
                {
                    target: "STORAGE",
                    exchange_buffer_id: in2BufferId,
                    is_in: true,
                    is_out: false,
                    width: -1,
                    height: -1
                },
                {
                    target: "STORAGE",
                    exchange_buffer_id: outBufferId,
                    is_in: false,
                    is_out: true,
                    width: -1,
                    height: -1
                },
                {
                    target: "STORAGE",
                    exchange_buffer_id: paramsBufferId,
                    is_in: true,
                    is_out: false,
                    width: -1,
                    height: -1
                }
            ]
        })
    )
    if (res < 0) {
        console.log("AN ERROR HAS OCCURED " + res)
    }


    moc.createImageFromRgbafloatPixels(textureWidth, textureHeight, textureBufferId, moc.getOutputBufferId())
    moc.writeExchangeBufferHeader(moc.getOutputBufferId(), "content-type", "image/png")

    return 200
}