#include "opengl.h"

GLuint renderHandle, computeHandle;

void updateTex(int);
void draw();

int main() {
	initGL();

	GLuint texHandle = genTexture();
	renderHandle = genRenderProg(texHandle);
	computeHandle = genComputeProg(texHandle);

	printf("ok 2\n");

	for (int i = 0; i < 1024; ++i) {
		updateTex(i);
		draw();
	}

	return 0;
}

void updateTex(int frame) {
	glUseProgram(computeHandle);
	glUniform1f(glGetUniformLocation(computeHandle, "roll"), (float)frame*0.01f);
	//glDispatchCompute(512/16, 512/16, 1); // 512^2 threads in blocks of 16^2
	printf("before dispatch\n");
	glDispatchCompute(1, 1, 1);
	printf("after dispatch\n");
	checkErrors("Dispatch compute shader");
}

void draw() {
	glUseProgram(renderHandle);
	glDrawArrays(GL_TRIANGLE_STRIP, 0, 4);
	swapBuffers();
	checkErrors("Draw screen");
}
