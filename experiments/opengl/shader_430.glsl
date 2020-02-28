#version 430

layout(std430) buffer;
layout (local_size_x = 1, local_size_y = 1) in;

layout(binding = 0) readonly buffer Input0 {
    float elements[];
} input_data0;

layout(binding = 1) readonly buffer Input1 {
    float elements[];
} input_data1;

layout(binding = 2) writeonly buffer Output {
    float elements[];
} output_data;

void main() {
    uint ident = gl_GlobalInvocationID.x;
    output_data.elements[ident] = input_data0.elements[ident] * input_data1.elements[ident];

    //ivec2 storePos = ivec2(gl_GlobalInvocationID.xy);
    //float localCoef = length(vec2(ivec2(gl_LocalInvocationID.xy)-8)/8.0);
    //float globalCoef = sin(float(gl_WorkGroupID.x+gl_WorkGroupID.y)*0.1 + roll)*0.5;
    //imageStore(destTex, storePos, vec4(1.0-globalCoef*localCoef, 0.0, 0.0, 0.0));
}
