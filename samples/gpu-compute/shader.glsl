#version 310 es

layout(local_size_x = 1) in;
layout(std430) buffer;

layout(binding = 0, rgba32f) writeonly uniform highp image2D uImage;

layout(binding = 1) readonly buffer Input0 {
    float elements[];
} input_data0;

layout(binding = 2) readonly buffer Input1 {
    float elements[];
} input_data1;

layout(binding = 3) writeonly buffer Output {
    float elements[];
} output_data;

layout(binding = 4) readonly buffer Params {
    float fRatio;
    float imageSize;
} parameters;

void main()
{
    float real = (-2.5+float(gl_GlobalInvocationID.x)*3.5)/parameters.imageSize;
    float imag = (-1.0+float(gl_GlobalInvocationID.y)*2.0)/parameters.imageSize;
    float cReal = real;
    float cImag = imag;

    float r2 = 0.0;
    int iter;

    int maxIterations = 25;
    for (iter = 0; iter < maxIterations && r2 < 4.0; ++iter)
    {
        float tempreal = real;
        real = (tempreal * tempreal) - (imag * imag) + cReal;
        imag = 2.0 * tempreal * imag + cImag;

        r2 = real*real + imag*imag;
    }

    vec4 color;
    if (r2 < 4.0)
        color = vec4(0.0, 0.0, 0.0, 0.0);
    else
        color = vec4(float(iter), 1.0, 2.0, 3.0);

    uint ident = gl_GlobalInvocationID.x;
    output_data.elements[ident] = parameters.fRatio + input_data0.elements[ident] * input_data1.elements[ident];

    imageStore(uImage, ivec2(gl_GlobalInvocationID.x,gl_GlobalInvocationID.y), color);
}
