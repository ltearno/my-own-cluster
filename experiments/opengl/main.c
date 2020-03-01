#include <assert.h>
#include <fcntl.h>
#include <stdbool.h>
#include <stdio.h>
#include <string.h>
#include <unistd.h>
#include <sys/stat.h>

#include <gbm.h>

#include <EGL/egl.h>
#include <EGL/eglext.h>

#include <GL/gl.h>
#include <GL/glext.h>

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

void checkErrors() {
	GLenum e = glGetError();
	if (e != GL_NO_ERROR) {
		//fprintf(stderr, "OpenGL error: %s (%d)\n", gluErrorString(e), e);
        fprintf(stderr, "OpenGL error: (%x)\n", e);
		exit(20);
	}
}

static const EGLint defaultConfigAttrs[] = {
   EGL_SURFACE_TYPE, EGL_PBUFFER_BIT, // EGL_WINDOW_BIT is the default and we don't want that on headless config
   /*EGL_BLUE_SIZE, 8,
   EGL_GREEN_SIZE, 8,
   EGL_RED_SIZE, 8,
   EGL_DEPTH_SIZE, 8,*/
   EGL_RENDERABLE_TYPE, EGL_OPENGL_ES3_BIT_KHR,
   EGL_NONE,
};

static const EGLint gbmConfigAttrs[] = {
   EGL_SURFACE_TYPE, EGL_WINDOW_BIT,
   /*EGL_BLUE_SIZE, 8,
   EGL_GREEN_SIZE, 8,
   EGL_RED_SIZE, 8,
   EGL_DEPTH_SIZE, 8,*/
   EGL_RENDERABLE_TYPE, EGL_OPENGL_ES3_BIT_KHR,
   EGL_NONE,
};

int main() {
   EGLDisplay egl_dpy;

   EGLint* configAttrs;

   if(getenv("DISPLAY")!=NULL) {
      int fd = open ("/dev/dri/renderD128", O_RDWR);
      assert (fd > 0);
      printf("opened dri device %d\n", fd);

      struct gbm_device *gbm = gbm_create_device (fd);
      assert (gbm != NULL);
      printf("opened gbm device %p\n", gbm);

      egl_dpy = eglGetPlatformDisplay (EGL_PLATFORM_GBM_KHR, gbm, NULL);

      configAttrs = gbmConfigAttrs;
   }
   else {
      egl_dpy = eglGetDisplay(EGL_DEFAULT_DISPLAY);

      configAttrs = defaultConfigAttrs;
   }

   assert (egl_dpy != NULL);
   printf("opened egl platform display %p\n", egl_dpy);

   EGLint major, minor;
   EGLBoolean res = eglInitialize (egl_dpy, &major, &minor);
   assert (res);
   printf("egl version %d.%d\n", major, minor);

   const char *egl_extension_st = eglQueryString (egl_dpy, EGL_EXTENSIONS);
   assert (strstr (egl_extension_st, "EGL_KHR_create_context") != NULL);
   assert (strstr (egl_extension_st, "EGL_KHR_surfaceless_context") != NULL);

   printf("EGL_CLIENT_APIS: '%s'\n", eglQueryString(egl_dpy, EGL_CLIENT_APIS));
   printf("EGL_EXTENSIONS: '%s'\n", eglQueryString(egl_dpy, EGL_EXTENSIONS));
   printf("EGL_VENDOR: '%s'\n", eglQueryString(egl_dpy, EGL_VENDOR));
   printf("EGL_VERSION: '%s'\n", eglQueryString(egl_dpy, EGL_VERSION));
   
   EGLint configCount;
   res = eglChooseConfig (egl_dpy, configAttrs, NULL, 0, &configCount);
   printf("config_count: %d\n", configCount);
   fflush(stdout);

   EGLConfig *configs = (EGLConfig*) malloc(sizeof(EGLConfig) * configCount);
   res = eglChooseConfig (egl_dpy, configAttrs, configs, configCount, &configCount);
   assert (res);

   // EGL_OPENGL_API, EGL_OPENGL_ES_API, or EGL_OPENVG_API
   res = eglBindAPI (EGL_OPENGL_API);
   assert (res);

   static const EGLint attribs[] = {
      //EGL_CONTEXT_CLIENT_VERSION, 3,
      EGL_NONE
   };

   EGLContext core_ctx = eglCreateContext (egl_dpy, configs[0], EGL_NO_CONTEXT, attribs);
   assert (core_ctx != EGL_NO_CONTEXT);
   printf("egl context created\n");
   // EGL_CONFIG_ID, EGL_CONTEXT_CLIENT_TYPE, EGL_CONTEXT_CLIENT_VERSION, or EGL_RENDER_BUFFER
   EGLint contextValue;
   res = eglQueryContext(egl_dpy, core_ctx, EGL_CONTEXT_CLIENT_TYPE, &contextValue);
   assert (res);
   printf("core_ctx EGL_CONTEXT_CLIENT_TYPE: %x\n", contextValue);
   assert (contextValue == EGL_OPENGL_API);
   res = eglQueryContext(egl_dpy, core_ctx, EGL_RENDER_BUFFER, &contextValue);
   assert (res);
   printf("core_ctx EGL_RENDER_BUFFER: %x = ", contextValue);
   switch(contextValue) {
      case EGL_SINGLE_BUFFER: printf("EGL_SINGLE_BUFFER\n"); break;
      case EGL_BACK_BUFFER: printf("EGL_BACK_BUFFER\n"); break;
      case EGL_NONE: printf("EGL_NONE"); break;
      default: printf("UNKNOWN"); break;
   };
   printf("\n");

   res = eglMakeCurrent (egl_dpy, EGL_NO_SURFACE, EGL_NO_SURFACE, core_ctx);
   assert (res);



   printf("GL_VERSION: '%s'\n", glGetString(GL_VERSION));
   printf("GL_VENDOR: '%s'\n", glGetString(GL_VENDOR));
   printf("GL_RENDERER: '%s'\n", glGetString(GL_RENDERER));
   printf("GL_EXTENSIONS: '%s'\n", glGetString(GL_EXTENSIONS));
   
   /* print some compute limits (not strictly necessary) */
   GLint work_group_count[3] = {0};
   for (unsigned i = 0; i < 3; i++)
      glGetIntegeri_v (GL_MAX_COMPUTE_WORK_GROUP_COUNT,
                       i,
                       &work_group_count[i]);
   printf ("GL_MAX_COMPUTE_WORK_GROUP_COUNT: %d, %d, %d\n",
           work_group_count[0],
           work_group_count[1],
           work_group_count[2]);

   GLint work_group_size[3] = {0};
   for (unsigned i = 0; i < 3; i++)
      glGetIntegeri_v (GL_MAX_COMPUTE_WORK_GROUP_SIZE, i, &work_group_size[i]);
   printf ("GL_MAX_COMPUTE_WORK_GROUP_SIZE: %d, %d, %d\n",
           work_group_size[0],
           work_group_size[1],
           work_group_size[2]);

   GLint max_invocations;
   glGetIntegerv (GL_MAX_COMPUTE_WORK_GROUP_INVOCATIONS, &max_invocations);
   printf ("GL_MAX_COMPUTE_WORK_GROUP_INVOCATIONS: %d\n", max_invocations);

   GLint mem_size;
   glGetIntegerv (GL_MAX_COMPUTE_SHARED_MEMORY_SIZE, &mem_size);
   printf ("GL_MAX_COMPUTE_SHARED_MEMORY_SIZE: %d\n", mem_size);
   
   /* setup a compute shader */
   printf("compute_shader creating...\n");
   GLuint compute_shader = glCreateShader (GL_COMPUTE_SHADER);
   checkErrors();
   printf("compute_shader created\n");

   const char *shader_source = loadText("shader_310.glsl");

   glShaderSource (compute_shader, 1, &shader_source, NULL);
   checkErrors();

   glCompileShader (compute_shader);
   GLint isCompiled = 0;
   glGetShaderiv(compute_shader, GL_COMPILE_STATUS, &isCompiled);
   if(isCompiled == GL_FALSE)
   {
      GLint maxLength = 0;
      glGetShaderiv(compute_shader, GL_INFO_LOG_LENGTH, &maxLength);

      // The maxLength includes the NULL character
      GLchar *errorLog = malloc(maxLength);
      glGetShaderInfoLog(compute_shader, maxLength, &maxLength, &errorLog[0]);

      printf("error: %s\n", errorLog);

      glDeleteShader(compute_shader);
      return;
   }
   checkErrors();

   GLuint shader_program = glCreateProgram ();

   glAttachShader (shader_program, compute_shader);
   checkErrors();

   glLinkProgram (shader_program);
   GLint rvalue;
   glGetProgramiv(shader_program, GL_LINK_STATUS, &rvalue);
   if (!rvalue) {
      fprintf(stderr, "Error in linking compute shader program\n");
      GLchar log[10240];
      GLsizei length;
      glGetProgramInfoLog(shader_program, 10239, &length, log);
      fprintf(stderr, "Linker log:\n%s\n", log);
      return;
   }   

   checkErrors();

   glDeleteShader (compute_shader);

   glUseProgram (shader_program);
   checkErrors();

   /* prepare input and output */
   const int dataSize = 4;
   float *in1 = malloc(sizeof(float) * dataSize);
   float *in2 = malloc(sizeof(float) * dataSize);
   for(int i=0;i<dataSize; i++ ){
      in1[i] = i;
      in2[i] = i;
   }

   GLuint textureIndex;
   int textureWidth = dataSize;
   int textureHeight = dataSize;
   glGenTextures(1, &textureIndex);
   glBindTexture(GL_TEXTURE_2D, textureIndex);
   glTexParameteri( GL_TEXTURE_2D, GL_TEXTURE_MIN_FILTER, GL_NEAREST );
   glTexParameteri( GL_TEXTURE_2D, GL_TEXTURE_MAG_FILTER, GL_NEAREST );
   //glTexImage2D(GL_TEXTURE_2D, 0, GL_RGBA32F, textureWidth, textureHeight, 0, GL_RGBA, GL_FLOAT, NULL);
   glTexStorage2D(GL_TEXTURE_2D, 1, GL_RGBA32F, textureWidth, textureHeight);
   checkErrors();
   glBindImageTexture(0, textureIndex, 0, GL_FALSE, 0, GL_WRITE_ONLY, GL_RGBA32F);
   glUniform1i(glGetUniformLocation(shader_program, "uImage"), 0);
   
   GLuint in1Index;
   glGenBuffers(1, &in1Index);
   glBindBuffer(GL_SHADER_STORAGE_BUFFER, in1Index);
   glBufferData(GL_SHADER_STORAGE_BUFFER, sizeof(float) * dataSize, in1,  GL_STATIC_DRAW);
   glBindBufferBase(GL_SHADER_STORAGE_BUFFER, 1, in1Index);
   
   GLuint in2Index;
   glGenBuffers(1, &in2Index);
   glBindBuffer(GL_SHADER_STORAGE_BUFFER, in2Index);
   glBufferData(GL_SHADER_STORAGE_BUFFER, sizeof(float) * dataSize, in2,  GL_STATIC_DRAW);
   glBindBufferBase(GL_SHADER_STORAGE_BUFFER, 2, in2Index);
   
   GLuint outIndex;
   glGenBuffers(1, &outIndex);
   glBindBuffer(GL_SHADER_STORAGE_BUFFER, outIndex);
   glBufferData(GL_SHADER_STORAGE_BUFFER, sizeof(float) * dataSize, NULL,  GL_STATIC_DRAW);
   glBindBufferBase(GL_SHADER_STORAGE_BUFFER, 3, outIndex);

   GLuint paramsIndex;
   glGenBuffers(1, &paramsIndex);
   glBindBuffer(GL_SHADER_STORAGE_BUFFER, paramsIndex);
   float fRatio = 1;
   glBufferData(GL_SHADER_STORAGE_BUFFER, sizeof(float) * 1, &fRatio,  GL_STATIC_DRAW);
   glBindBufferBase(GL_SHADER_STORAGE_BUFFER, 4, paramsIndex);
   
   checkErrors();
   printf("buffers %d %d %d %d\n", in1Index, in2Index, outIndex, paramsIndex);

   /* dispatch computation */
   glDispatchCompute (dataSize, dataSize, 1);
   checkErrors();

   //glMemoryBarrier(GL_BUFFER_UPDATE_BARRIER_BIT);
   glMemoryBarrier(GL_ALL_BARRIER_BITS);
   checkErrors();

   printf ("Compute shader dispatched and finished successfully\n");

   glBindBuffer(GL_SHADER_STORAGE_BUFFER, outIndex);
   checkErrors();
   float* tmp = malloc(sizeof(float) * dataSize);
   tmp[0] = 42;
   glGetBufferSubData(GL_SHADER_STORAGE_BUFFER, 0, sizeof(float) * dataSize, tmp);
   printf("tmp buffer: %p\n", tmp);
   for(int i=0;i<10;i++)
      printf("tmp[%d] = %f\n", i, tmp[i]);
   
   glBindTexture(GL_TEXTURE_2D, textureIndex);
   float *pixels = malloc(sizeof(float)*textureWidth*textureHeight*4);
   glGetTexImage(GL_TEXTURE_2D, 0, GL_RGBA, GL_FLOAT, pixels);
   for(int x=0;x<textureWidth*textureHeight*4;x+=4)
      printf("%d: %f %f %f %f\n",x,  pixels[x], pixels[x+1], pixels[x+2], pixels[x+3]);

   /* free stuff */
   glDeleteProgram (shader_program);
   //eglDestroyContext (egl_dpy, core_ctx);
   eglTerminate (egl_dpy);
   //gbm_device_destroy (gbm);
   //close (fd);
}