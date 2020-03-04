#version 310 es

layout(local_size_x = 256) in;
layout(std430) buffer;

layout(binding = 0, rgba32f) writeonly uniform highp image2D uImage;

layout(binding = 1) readonly buffer Params {
    float xmin;
    float xspan;
    float ymin;
    float yspan;
    float imageSize;
    float maxIterations;
    float mult;
} parameters;

void main()
{
    float real = parameters.xmin + (float(gl_GlobalInvocationID.x)*parameters.xspan)/parameters.imageSize;
    float imag = parameters.ymin + (float(gl_GlobalInvocationID.y)*parameters.yspan)/parameters.imageSize;

    float cReal = real;
    float cImag = imag;

    float r2 = 0.0;
    float iter;
    float maxIterations = parameters.maxIterations;

    for (iter = 0.0; iter < maxIterations && r2 < 4.0; iter+=1.0)
    {
        float tempreal = real;
        real = (tempreal * tempreal) - (imag * imag) + cReal;
        imag = 2.0 * tempreal * imag + cImag;

        r2 = real*real + imag*imag;
    }

    vec4 color;
    if (r2 < 4.0)
        color = vec4(0.0, 0.0, 0.0, 1.0);
    else
        color = vec4(parameters.mult * float(iter)/float(maxIterations), 0.0, 0.0, 1.0);

    imageStore(uImage, ivec2(gl_GlobalInvocationID.x,gl_GlobalInvocationID.y), color);
}
