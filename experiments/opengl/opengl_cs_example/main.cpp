#include "opengl.h"

GLuint computeHandle;

void updateTex(int);
void draw();

int main() {
	initGL();

	GLuint texHandle = genTexture();
	computeHandle = genComputeProg(texHandle);

	printf("ok 2\n");

	for (int i = 0; i < 1024; ++i) {
		updateTex(i);
	}

	return 0;
}

void updateTex(int frame) {
	glUseProgram(computeHandle);
	glUniform1f(glGetUniformLocation(computeHandle, "roll"), (float)frame*0.01f);
	glDispatchCompute(512/16, 512/16, 1); // 512^2 threads in blocks of 16^2
	checkErrors("Dispatch compute shader");
}
