package main

/*
#cgo LDFLAGS: -lOpenGL -lGLU -lEGL
#define GL_GLEXT_PROTOTYPES
#include <fcntl.h>
#include <EGL/egl.h>
#include <EGL/eglext.h>
#include <GL/gl.h>
#include <sys/stat.h>
#include <stdlib.h>

#include "glext.h"

#include <stdio.h>

void err(int n, char* s){
	printf("err %d %s\n", n, s);
	exit(n);
}

void checkErrors(const char* when) {
	GLenum e = glGetError();
	if (e != GL_NO_ERROR) {
		//fprintf(stderr, "OpenGL error: %s (%d)\n", gluErrorString(e), e);
        fprintf(stderr, "OpenGL error when %s: (0x%x)\n", when, e);
		exit(20);
	}
}

  static const EGLint configAttribs[] = {
          EGL_SURFACE_TYPE, EGL_PBUFFER_BIT,
          EGL_BLUE_SIZE, 8,
          EGL_GREEN_SIZE, 8,
          EGL_RED_SIZE, 8,
          EGL_DEPTH_SIZE, 8,
          EGL_RENDERABLE_TYPE, EGL_OPENGL_BIT,
          EGL_NONE
  };

  static const EGLint pbufferAttribs[] = {
        EGL_WIDTH, 9,
        EGL_HEIGHT, 9,
        EGL_NONE,
  };

int runOpenGLDemo(const char *shader_source)
{
  // 1. Initialize EGL
  unsetenv("DISPLAY");
  EGLDisplay eglDpy = eglGetDisplay(EGL_DEFAULT_DISPLAY);

  printf("EGL_CLIENT_APIS: '%s'\n", eglQueryString(eglDpy, EGL_CLIENT_APIS));
  printf("EGL_EXTENSIONS: '%s'\n", eglQueryString(eglDpy, EGL_EXTENSIONS));
  printf("EGL_VENDOR: '%s'\n", eglQueryString(eglDpy, EGL_VENDOR));
  printf("EGL_VERSION: '%s'\n", eglQueryString(eglDpy, EGL_VERSION));

  EGLint major, minor;
  eglInitialize(eglDpy, &major, &minor);
  printf("egl version %d.%d\n", major, minor);

  // 2. Select an appropriate configuration
  EGLint configCount;
  eglChooseConfig (eglDpy, configAttribs, NULL, 0, &configCount);
  printf("config_count: %d\n", configCount);

  EGLConfig *eglConfigs = (EGLConfig*) malloc(sizeof(EGLConfig) * configCount);
  eglChooseConfig (eglDpy, configAttribs, eglConfigs, configCount, &configCount);

  EGLConfig eglCfg = eglConfigs[0];
  //eglChooseConfig(eglDpy, configAttribs, &eglCfg, 1, &configCount);

  // 3. Create a surface
  EGLSurface eglSurf = eglCreatePbufferSurface(eglDpy, eglCfg, pbufferAttribs);

  // 4. Bind the API
  eglBindAPI(EGL_OPENGL_API);

  // 5. Create a context and make it current
  EGLint eglAttrs[] = {
    //EGL_CONTEXT_CLIENT_VERSION, 4,
    //EGL_CONTEXT_MINOR_VERSION, 6,
    EGL_NONE
  };

  EGLContext eglCtx = eglCreateContext(eglDpy, eglCfg, EGL_NO_CONTEXT, eglAttrs);

  eglMakeCurrent(eglDpy, eglSurf, eglSurf, eglCtx);

  printf("GL_VERSION: '%s'\n", glGetString(GL_VERSION));
  printf("GL_VENDOR: '%s'\n", glGetString(GL_VENDOR));
  printf("GL_RENDERER: '%s'\n", glGetString(GL_RENDERER));
  printf("GL_EXTENSIONS: '%s'\n", glGetString(GL_EXTENSIONS));

  // prepare input and output
  const int dataSize = 1024;
  float *in1 = malloc(sizeof(float) * dataSize);
  float *in2 = malloc(sizeof(float) * dataSize);
  for(int i=0;i<dataSize; i++ ){
    in1[i] = in2[i] = i;
  }
  GLuint in1Index, in2Index, outIndex;
  glGenBuffers(1, &in1Index);
  checkErrors("gen buffers");
  glBindBuffer(GL_SHADER_STORAGE_BUFFER, in1Index);
  checkErrors("bind buffer");
  glBufferData(GL_SHADER_STORAGE_BUFFER, sizeof(float) * dataSize, in1, GL_DYNAMIC_COPY);
  checkErrors("buffer data");
  glBindBufferBase(GL_SHADER_STORAGE_BUFFER, 0, in1Index);
  checkErrors("bind buffer base");
  glGenBuffers(1, &in2Index);
  glBindBuffer(GL_SHADER_STORAGE_BUFFER, in2Index);
  glBufferData(GL_SHADER_STORAGE_BUFFER, sizeof(float) * dataSize, in2, GL_DYNAMIC_COPY);
  glBindBufferBase(GL_SHADER_STORAGE_BUFFER, 1, in2Index);
  glGenBuffers(1, &outIndex);
  glBindBuffer(GL_SHADER_STORAGE_BUFFER, outIndex);
  glBufferData(GL_SHADER_STORAGE_BUFFER, sizeof(float) * dataSize, NULL, GL_DYNAMIC_COPY);
  glBindBufferBase(GL_SHADER_STORAGE_BUFFER, 2, outIndex);
  checkErrors("creating and binding buffers");

  printf("buffers: %d %d %d\n", in1Index, in2Index, outIndex);

  // setup a compute shader
  printf("compute_shader creating...\n");
  GLuint compute_shader = glCreateShader (GL_COMPUTE_SHADER);
  checkErrors("creating shader");
  printf("compute_shader created\n");

  glShaderSource (compute_shader, 1, &shader_source, NULL);
  checkErrors("giving shader source");

  glCompileShader (compute_shader);
  checkErrors("compiling shader");

  GLuint shader_program = glCreateProgram ();

  glAttachShader (shader_program, compute_shader);
  checkErrors("attaching shader");

  glLinkProgram (shader_program);
  checkErrors("linking shader");

  glDeleteShader (compute_shader);

  glUseProgram (shader_program);
  checkErrors("using shader in program");

  // dispatch computation
  glDispatchCompute (dataSize, 1, 1);
  checkErrors("dispatching compute");

  glBindBuffer(GL_SHADER_STORAGE_BUFFER, outIndex);
  checkErrors("binding buffer");
  float* tmp = malloc(sizeof(float) * dataSize);
  glGetBufferSubData(GL_SHADER_STORAGE_BUFFER, 0, sizeof(float) * dataSize, tmp);
  checkErrors("getting buffer sub data");
  printf("tmp buffer: %p\n", tmp);
  for(int i=0;i<10;i++)
    printf("tmp[%d] = %f\n", i, tmp[i]);

  // 6. Terminate EGL when finished
  eglTerminate(eglDpy);
  printf("ok\n");
  return 0;
}
*/
import "C"

var shaderCode string = `
#version 430

layout(std430) buffer;
layout (local_size_x = 1, local_size_y = 1, local_size_z = 1) in;

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
`

func TestOpenGL() {
	C.runOpenGLDemo(C.CString(shaderCode))
}
