#include <fcntl.h>
#include <EGL/egl.h>
#include <EGL/eglext.h>
#include <GL/gl.h>
#include <sys/stat.h>

#include <stdio.h>

void checkErrors(const char* when) {
	GLenum e = glGetError();
	if (e != GL_NO_ERROR) {
		//fprintf(stderr, "OpenGL error: %s (%d)\n", gluErrorString(e), e);
        fprintf(stderr, "OpenGL error when %s: (0x%x)\n", when, e);
		exit(20);
	}
}

// loads a binary file, allocate and return the file content
char* loadText(char *fileName) {
    int fd = open(fileName, O_RDONLY);
    if (fd < 0)
        err(1, "can not open binary file\n");

    struct stat stat;
    fstat(fd, &stat);

    unsigned char* buffer = malloc(stat.st_size +1);

    int readden = read(fd, buffer, stat.st_size);
    buffer[stat.st_size] = 0;
    printf("read size: %d\n", readden);

    return buffer;
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

  static const int pbufferWidth = 9;
  static const int pbufferHeight = 9;

  static const EGLint pbufferAttribs[] = {
        EGL_WIDTH, pbufferWidth,
        EGL_HEIGHT, pbufferHeight,
        EGL_NONE,
  };

int main(int argc, char *argv[])
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
  /*for(int i=0; i<configCount; i++) {
    printf("config %d\n", i);
    EGLint configValue;
    eglGetConfigAttrib(eglDpy, eglConfigs[i], EGL_RENDERABLE_TYPE, &configValue);
    printf("EGL_RENDERABLE_TYPE: %d\n", configValue);
  }*/

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

  /* prepare input and output */
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
  
  /* setup a compute shader */
  printf("compute_shader creating...\n");
  GLuint compute_shader = glCreateShader (GL_COMPUTE_SHADER);
  checkErrors("creating shader");
  printf("compute_shader created\n");

  const char *shader_source = loadText("shader_430.glsl");

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

  /* dispatch computation */
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

  /*float *outBound = (float*) glMapBuffer(GL_SHADER_STORAGE_BUFFER, GL_READ_ONLY);
  checkErrors("mapping buffer");
  printf("outBound buffer: %p\n", outBound);
  for(int i=8;i<10;i++)
  printf("outBound[%d] = %f\n", i, outBound[i]);
  glUnmapBuffer(GL_SHADER_STORAGE_BUFFER);*/

  // 6. Terminate EGL when finished
  eglTerminate(eglDpy);
  printf("ok\n");
  return 0;
}