#include <fcntl.h>
#include <EGL/egl.h>
#include <GL/gl.h>
#include <sys/stat.h>

#include <stdio.h>

void checkErrors() {
	GLenum e = glGetError();
	if (e != GL_NO_ERROR) {
		//fprintf(stderr, "OpenGL error: %s (%d)\n", gluErrorString(e), e);
        fprintf(stderr, "OpenGL error: (%x)\n", e);
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
  EGLDisplay eglDpy = eglGetDisplay(EGL_DEFAULT_DISPLAY);

  printf("EGL_CLIENT_APIS: '%s'\n", eglQueryString(eglDpy, EGL_CLIENT_APIS));
   printf("EGL_EXTENSIONS: '%s'\n", eglQueryString(eglDpy, EGL_EXTENSIONS));
   printf("EGL_VENDOR: '%s'\n", eglQueryString(eglDpy, EGL_VENDOR));
   printf("EGL_VERSION: '%s'\n", eglQueryString(eglDpy, EGL_VERSION));

  EGLint major, minor;
  eglInitialize(eglDpy, &major, &minor);
  printf("egl version %d.%d\n", major, minor);

  // 2. Select an appropriate configuration
  EGLint numConfigs;
  EGLConfig eglCfg;

  eglChooseConfig(eglDpy, configAttribs, &eglCfg, 1, &numConfigs);

  // 3. Create a surface
  EGLSurface eglSurf = eglCreatePbufferSurface(eglDpy, eglCfg, pbufferAttribs);

  // 4. Bind the API
  eglBindAPI(EGL_OPENGL_API);

  // 5. Create a context and make it current
  EGLContext eglCtx = eglCreateContext(eglDpy, eglCfg, EGL_NO_CONTEXT, NULL);

  eglMakeCurrent(eglDpy, eglSurf, eglSurf, eglCtx);


  

  // from now on use your OpenGL context

  printf("GL_VERSION: '%s'\n", glGetString(GL_VERSION));
   printf("GL_VENDOR: '%s'\n", glGetString(GL_VENDOR));
   printf("GL_RENDERER: '%s'\n", glGetString(GL_RENDERER));
   printf("GL_EXTENSIONS: '%s'\n", glGetString(GL_EXTENSIONS));




   /* prepare input and output */
   /*
   const int dataSize = 1024;
   float *in1 = malloc(sizeof(float) * dataSize);
   float *in2 = malloc(sizeof(float) * dataSize);
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
   checkErrors();

   printf("buffers: %d %d %d\n", in1Index, in2Index, outIndex);


   glBindBuffer(GL_SHADER_STORAGE_BUFFER, outIndex);
   checkErrors();
   float *outBound = (float*) glMapBuffer(GL_SHADER_STORAGE_BUFFER, GL_READ_ONLY);
   checkErrors();
   printf("outBound buffer: %p\n", outBound);
   for(int i=8;i<10;i++)
      printf("outBound[%d] = %f\n", i, outBound[i]);
   glUnmapBuffer(GL_SHADER_STORAGE_BUFFER);
*/



  /* setup a compute shader */
   printf("compute_shader creating...\n");
   GLuint compute_shader = glCreateShader (GL_COMPUTE_SHADER);
   checkErrors();
   printf("compute_shader created\n");

   const char *shader_source = loadText("shader_101.glsl");

   glShaderSource (compute_shader, 1, &shader_source, NULL);
   checkErrors();

   glCompileShader (compute_shader);
   checkErrors();

   GLuint shader_program = glCreateProgram ();

   glAttachShader (shader_program, compute_shader);
   checkErrors();

   glLinkProgram (shader_program);
   checkErrors();

   glDeleteShader (compute_shader);

   glUseProgram (shader_program);
   checkErrors();




  // 6. Terminate EGL when finished
  eglTerminate(eglDpy);
  printf("ok\n");
  return 0;
}