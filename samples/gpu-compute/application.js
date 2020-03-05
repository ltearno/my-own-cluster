/// reference path="./core-api-guest.d.ts"
/// reference path="./gpu-api-guest.d.ts"

const moc = requireApi('core')
const gpu = requireApi('gpu')

function getRequestParameters() {
    var bufferId = moc.getInputBufferId();
    var headers = moc.readExchangeBufferHeaders(bufferId);
    var query = headers["x-moc-url-query"] || "{}";
    query = decodeURI(query);
    var params = JSON.parse(query) || {};
    params.xmin = params.xmin || 0;
    params.xmax = params.xmax || 0;
    params.ymin = params.ymin || 0;
    params.ymax = params.ymax || 0;
    params.iter = params.iter || 12;
    params.mult = params.mult || 10;
    return params
}

function launchMandelbrotShader() {
    var req = getRequestParameters()

    console.log("processing query " + JSON.stringify(req))

    var dataSize = 1024;
    var textureWidth = dataSize;
    var textureHeight = dataSize;
    
    var texture = new Float32Array(4 * textureWidth * textureHeight);
    var textureBufferId = moc.createExchangeBuffer();
    // we write in the buffer because opengl api will read it to fill the texture (no option yet)
    moc.writeExchangeBuffer(textureBufferId, texture);

    var params = new Float32Array(7)
    params[0] = req.xmin;
    params[1] = req.xmax - req.xmin;
    params[2] = req.ymin;
    params[3] = req.ymax - req.ymin;
    params[4] = dataSize;
    params[5] = req.iter;
    params[6] = req.mult;
    var paramsBufferId = moc.createExchangeBuffer();
    moc.writeExchangeBuffer(paramsBufferId, params);

    var res = gpu.computeShader(
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
        console.log("error occured when processing shader" + res)
    }

    gpu.createImageFromRgbaFloatPixels(textureWidth, textureHeight, textureBufferId, moc.getOutputBufferId())
    moc.writeExchangeBufferHeader(moc.getOutputBufferId(), "content-type", "image/png")

    return 200
}

function test() {
    var dataSize = 10;
    var values = new Float32Array(dataSize);
    for (var i = 0; i < dataSize; i++) {
        values[i] = i;
    }

    var bufferId = moc.createExchangeBuffer();
    var written = moc.writeExchangeBuffer(bufferId, values);
    console.log("written " + written + " values")

    console.log("yoiii " + bufferId)
    var bytes = moc.readExchangeBuffer(bufferId);
    console.log("yo")
    var outValues = new Float32Array(bytes.buffer);
    for (var i = 0; i < dataSize; i++)
        console.log(outValues[i] + " / " + values[i])
}