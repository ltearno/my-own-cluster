#version 430

layout(std430) buffer;
layout(local_size_x = 1, local_size_y = 1, local_size_z = 1) in;

// bound to input exchange buffer
layout(binding = 0) readonly  buffer Input0 { float elements[]; } input_data0;

// bound to output exchange buffer
layout(binding = 1) writeonly buffer Output { float elements[]; } output_data;

void main() {
    uint ident = gl_GlobalInvocationID.x;

    // simply square the input
    output_data.elements[ident] = ident + input_data0.elements[ident] * input_data0.elements[ident];
}
