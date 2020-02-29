#version 310 es

layout(local_size_x = 1) in;
layout(std430) buffer;

layout(binding = 0) readonly buffer Input0 {
    float elements[];
} input_data0;

layout(binding = 1) readonly buffer Input1 {
    float elements[];
} input_data1;

layout(binding = 2) writeonly buffer Output {
    float elements[];
} output_data;

layout(binding = 3) readonly buffer Params {
    float fRatio;
} parameters;

void main()
{
    uint ident = gl_GlobalInvocationID.x;
    output_data.elements[ident] = parameters.fRatio + input_data0.elements[ident] * input_data1.elements[ident];
}
