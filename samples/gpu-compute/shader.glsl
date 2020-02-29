#version 310 es

layout(local_size_x = 1) in;
layout(std430) buffer;

layout(binding = 0) readonly buffer Input0 {
    float elements[];
} input_data0;

layout(binding = 1) writeonly buffer Output {
    float elements[];
} output_data;

void main()
{
    uint ident = gl_GlobalInvocationID.x;
    output_data.elements[ident] = input_data0.elements[ident] * input_data0.elements[ident];
}
