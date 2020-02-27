#include <iostream>
#include <fstream>
using namespace std;

#include "opengl.h"

void checkErrors() {
	GLenum e = glGetError();
	if (e != GL_NO_ERROR) {
		//fprintf(stderr, "OpenGL error: %s (%d)\n", gluErrorString(e), e);
        fprintf(stderr, "OpenGL error: (%x)\n", e);
		exit(20);
	}
}

// loads a binary file, allocate and return the file content
const char* loadText(const char *fileName) {
   printf("a %s\n", fileName);
   ifstream in;
   in.open(fileName);

   std::string contents((std::istreambuf_iterator<char>(in)), 
    std::istreambuf_iterator<char>());

    cout << contents;
    cout.flush();

    return contents.c_str();
}

GLuint computeHandle;

void updateTex(int);
void draw();

int main() {
   initGL();

/* setup a compute shader */
   printf("compute_shader creating...\n");
   GLuint compute_shader = glCreateShader (GL_COMPUTE_SHADER);
   checkErrors();
   printf("compute_shader created\n");

   const char *shader_source = (const char *) loadText("shader_430.glsl");

   glShaderSource (compute_shader, 1, &shader_source, NULL);
   checkErrors();

   glCompileShader (compute_shader);
   int rvalue;
    glGetShaderiv(compute_shader, GL_COMPILE_STATUS, &rvalue);
    if (!rvalue) {
        fprintf(stderr, "Error in compiling the compute shader\n");
        GLchar log[10240];
        GLsizei length;
        glGetShaderInfoLog(compute_shader, 10239, &length, log);
        fprintf(stderr, "Compiler log:\n%s\n", log);
        exit(40);
    }
   checkErrors();

   GLuint shader_program = glCreateProgram ();

   glAttachShader (shader_program, compute_shader);
   checkErrors();

   glLinkProgram (shader_program);
   checkErrors();

   glDeleteShader (compute_shader);

   glUseProgram (shader_program);
   checkErrors();
   printf("ok shader created\n");

   /* prepare input and output */
   /*const int dataSize = 1;
   float *in1 = new float[dataSize];
   float *in2 = new float[dataSize];
   for(int i=0;i<dataSize; i++ ){
      in1[i] = in2[i] = i;
   }
   GLuint in1Index, in2Index, outIndex;
   glGenBuffers(1, &in1Index);
   glBindBuffer(GL_SHADER_STORAGE_BUFFER, in1Index);
   glBufferData(GL_SHADER_STORAGE_BUFFER, sizeof(float) * dataSize, in1, GL_DYNAMIC_COPY);
   glBindBufferBase(GL_SHADER_STORAGE_BUFFER, 0, in1Index);
   glGenBuffers(1, &in2Index);
   glBindBuffer(GL_SHADER_STORAGE_BUFFER, in2Index);
   glBufferData(GL_SHADER_STORAGE_BUFFER, sizeof(float) * dataSize, in2, GL_DYNAMIC_COPY);
   glBindBufferBase(GL_SHADER_STORAGE_BUFFER, 1, in2Index);
   glGenBuffers(1, &outIndex);
   glBindBuffer(GL_SHADER_STORAGE_BUFFER, outIndex);
   glBufferData(GL_SHADER_STORAGE_BUFFER, sizeof(float) * dataSize, NULL, GL_DYNAMIC_COPY);
   glBindBufferBase(GL_SHADER_STORAGE_BUFFER, 2, outIndex);
   checkErrors();*/

   /* dispatch computation */
   //glDispatchCompute (dataSize, 1, 1);
   //checkErrors();

   //glMemoryBarrier(GL_VERTEX_ATTRIB_ARRAY_BARRIER_BIT);
   //checkErrors();

   /*glBindBuffer(GL_SHADER_STORAGE_BUFFER, outIndex);
   checkErrors();
   float *outBound = (float*) glMapBuffer(GL_SHADER_STORAGE_BUFFER, GL_READ_ONLY);
   checkErrors();
   printf("outBound buffer: %p\n", outBound);
   for(int i=8;i<10;i++)
      printf("outBound[%d] = %f\n", i, outBound[i]);
   glUnmapBuffer(GL_SHADER_STORAGE_BUFFER);*/

   printf ("Compute shader dispatched and finished successfully\n");
}

int maddin() {
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
