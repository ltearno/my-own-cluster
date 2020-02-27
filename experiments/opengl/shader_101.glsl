#version 310 es
layout(local_size_x = 128) in;
layout(std430) buffer;
layout(binding = 0) writeonly buffer Output {
    vec4 elements[];
} output_data;
layout(binding = 1) readonly buffer Input0 {
    vec4 elements[];
} input_data0;
layout(binding = 2) readonly buffer Input1 {
    vec4 elements[];
} input_data1;
void main()
{
    uint ident = gl_GlobalInvocationID.x;
    output_data.elements[ident] = input_data0.elements[ident] * input_data1.elements[ident];
}
