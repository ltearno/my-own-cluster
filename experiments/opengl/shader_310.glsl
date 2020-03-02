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
    /*
    for each pixel (Px, Py) on the screen do
        x0 = scaled x coordinate of pixel (scaled to lie in the Mandelbrot X scale (-2.5, 1))
        y0 = scaled y coordinate of pixel (scaled to lie in the Mandelbrot Y scale (-1, 1))
        x := 0.0
        y := 0.0
        iteration := 0
        max_iteration := 1000
        while (x×x + y×y ≤ 2×2 AND iteration < max_iteration) do
            xtemp := x×x - y×y + x0
            y := 2×x×y + y0
            x := xtemp
            iteration := iteration + 1
    
        color := palette[iteration]
        plot(Px, Py, color)
    */

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
        color = vec4(0.0);
    else
        color = vec4(float(iter));

    uint ident = gl_GlobalInvocationID.x;
    output_data.elements[ident] = parameters.fRatio + input_data0.elements[ident] * input_data1.elements[ident];

    imageStore(uImage, ivec2(gl_GlobalInvocationID.x,gl_GlobalInvocationID.y), color);
}
