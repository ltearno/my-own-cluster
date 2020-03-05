#include "engine-opengl.h"

int checkErrors() {
	GLenum e = glGetError();
	if (e != GL_NO_ERROR) {
		fprintf(stderr, "OpenGL error: (%x)\n", e);
		return -1;
  }

  return 0;
}


static EGLint defaultConfigAttrs[] = {
   EGL_SURFACE_TYPE, EGL_PBUFFER_BIT, // EGL_WINDOW_BIT is the default and we don't want that on headless config
   EGL_RENDERABLE_TYPE, EGL_OPENGL_ES3_BIT_KHR,
   EGL_NONE,
};

static EGLint gbmConfigAttrs[] = {
   EGL_SURFACE_TYPE, EGL_WINDOW_BIT,
   EGL_RENDERABLE_TYPE, EGL_OPENGL_ES3_BIT_KHR,
   EGL_NONE,
};

static const EGLint contextAttribList[] = {
   EGL_NONE
};

void displayContextInfo(EGLDisplay egl_dpy, EGLContext core_ctx) {
   // EGL_CONFIG_ID, EGL_CONTEXT_CLIENT_TYPE, EGL_CONTEXT_CLIENT_VERSION, or EGL_RENDER_BUFFER
   EGLint contextValue;
   EGLBoolean res = eglQueryContext(egl_dpy, core_ctx, EGL_CONTEXT_CLIENT_TYPE, &contextValue);
   if (res)
      printf("core_ctx EGL_CONTEXT_CLIENT_TYPE: %x\n", contextValue);
   res = eglQueryContext(egl_dpy, core_ctx, EGL_RENDER_BUFFER, &contextValue);
   if (res);
      printf("core_ctx EGL_RENDER_BUFFER: %x = ", contextValue);
   switch(contextValue) {
      case EGL_SINGLE_BUFFER: printf("EGL_SINGLE_BUFFER\n"); break;
      case EGL_BACK_BUFFER: printf("EGL_BACK_BUFFER\n"); break;
      case EGL_NONE: printf("EGL_NONE"); break;
      default: printf("UNKNOWN"); break;
   };
   printf("\n");

   printf("GL_VERSION: '%s'\n", glGetString(GL_VERSION));
   printf("GL_VENDOR: '%s'\n", glGetString(GL_VENDOR));
   printf("GL_RENDERER: '%s'\n", glGetString(GL_RENDERER));
   printf("GL_EXTENSIONS: '%s'\n", glGetString(GL_EXTENSIONS));

   // print some compute limits
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
}

int initOpenGLDevice(DeviceInformation *info) {
   info->egl_dpy = NULL;
   info->gbm = NULL;
   info->renderDeviceFd = 0;

   EGLint* configAttrs;

   if(getenv("DISPLAY")!=NULL) {
      info->renderDeviceFd = open ("/dev/dri/renderD128", O_RDWR);
      if (info->renderDeviceFd <= 0)
         return -1;

      printf("opened dri device %d\n", info->renderDeviceFd);

      info->gbm = gbm_create_device (info->renderDeviceFd);
      if (info->gbm == NULL)
         return -2;

      printf("opened gbm device %p\n", info->gbm);

      info->egl_dpy = eglGetPlatformDisplay (EGL_PLATFORM_GBM_KHR, info->gbm, NULL);

      configAttrs = gbmConfigAttrs;
   }
   else {
      info->egl_dpy = eglGetDisplay(EGL_DEFAULT_DISPLAY);

      configAttrs = defaultConfigAttrs;
   }

   if (info->egl_dpy == NULL)
      return -7;

   printf("opened egl platform display %p\n", info->egl_dpy);

   EGLint major, minor;
   EGLBoolean res = eglInitialize (info->egl_dpy, &major, &minor);
   if (! res)
      return -4;

   printf("egl version %d.%d\n", major, minor);

   const char *egl_extension_st = eglQueryString (info->egl_dpy, EGL_EXTENSIONS);
   if (strstr (egl_extension_st, "EGL_KHR_create_context") == NULL)
      return -5;
   if (strstr (egl_extension_st, "EGL_KHR_surfaceless_context") == NULL)
      return -6;

   printf("EGL_CLIENT_APIS: '%s'\n", eglQueryString(info->egl_dpy, EGL_CLIENT_APIS));
   printf("EGL_EXTENSIONS: '%s'\n", eglQueryString(info->egl_dpy, EGL_EXTENSIONS));
   printf("EGL_VENDOR: '%s'\n", eglQueryString(info->egl_dpy, EGL_VENDOR));
   printf("EGL_VERSION: '%s'\n", eglQueryString(info->egl_dpy, EGL_VERSION));

   EGLint configCount;
   res = eglChooseConfig (info->egl_dpy, configAttrs, &info->config, 1, &configCount);
   if (!res)
      return -8;

   return 0;
}

int initOpenGLContext(DeviceInformation *dInfo, ContextInformation *info) {
   // EGL_OPENGL_API, EGL_OPENGL_ES_API, or EGL_OPENVG_API
   EGLBoolean res = eglBindAPI (EGL_OPENGL_API);
   if (!res)
      return -9;

   info->core_ctx = eglCreateContext (dInfo->egl_dpy, dInfo->config, EGL_NO_CONTEXT, contextAttribList);
   if (info->core_ctx == EGL_NO_CONTEXT)
      return -10;

   //printf("egl context created\n");

   return 0;
}

int setContext(DeviceInformation *dInfo, ContextInformation *info){
  int res = eglMakeCurrent (dInfo->egl_dpy, EGL_NO_SURFACE, EGL_NO_SURFACE, info->core_ctx);
  if (!res)
    return -11;
  return 0;
}

int clearContext(DeviceInformation *dInfo, ContextInformation *info){
   eglDestroyContext (dInfo->egl_dpy, info->core_ctx);
  int res = eglMakeCurrent (dInfo->egl_dpy, EGL_NO_SURFACE, EGL_NO_SURFACE, NULL);
  if (!res)
    return -11;
  return 0;
}

int destroyOpenGLContext(DeviceInformation *dInfo, ContextInformation *ctx) {
   eglDestroyContext (dInfo->egl_dpy, ctx->core_ctx);
   eglTerminate (dInfo->egl_dpy);
   if(dInfo->gbm != NULL){
      gbm_device_destroy (dInfo->gbm);
      dInfo->gbm = NULL;
   }
   if(dInfo->renderDeviceFd > 0) {
      close(dInfo->renderDeviceFd);
      dInfo->renderDeviceFd = 0;
   }
}

int compileAndBindShader(const char *shader_source) {
  //printf("compute_shader creating...\n");
   GLuint compute_shader = glCreateShader (GL_COMPUTE_SHADER);
   if(checkErrors()<0)
    return -1;
   //printf("compute_shader created\n");

   glShaderSource (compute_shader, 1, &shader_source, NULL);
   if(checkErrors()<0)
    return -1;

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
      return -2;
   }
   if(checkErrors()<0)
    return -1;

   GLuint shader_program = glCreateProgram ();

   glAttachShader (shader_program, compute_shader);
   if(checkErrors()<0)
    return -1;

   glLinkProgram (shader_program);
   GLint rvalue;
   glGetProgramiv(shader_program, GL_LINK_STATUS, &rvalue);
   if (!rvalue) {
      fprintf(stderr, "Error in linking compute shader program\n");
      GLchar log[10240];
      GLsizei length;
      glGetProgramInfoLog(shader_program, 10239, &length, log);
      fprintf(stderr, "Linker log:\n%s\n", log);
      return -3;
   }

   if(checkErrors()<0)
    return -1;

   glDeleteShader (compute_shader);

   glUseProgram (shader_program);
   if(checkErrors()<0)
    return -1;

   return 0;
}

int bindStorageBuffer(int binding, void* buffer, int size) {
  GLuint bufferIndex;

  glGenBuffers(1, &bufferIndex);
  glBindBuffer(GL_SHADER_STORAGE_BUFFER, bufferIndex);
  glBufferData(GL_SHADER_STORAGE_BUFFER, size, buffer,  GL_STATIC_DRAW);
  glBindBufferBase(GL_SHADER_STORAGE_BUFFER, binding, bufferIndex);

  return bufferIndex;
}

int bindTexture2DRGBAFloat(int binding, int textureWidth, int textureHeight) {
  int dataSize = sizeof(float) * textureWidth * textureHeight;

  GLuint textureIndex;

  glGenTextures(1, &textureIndex);
  glBindTexture(GL_TEXTURE_2D, textureIndex);
  glTexParameteri( GL_TEXTURE_2D, GL_TEXTURE_MIN_FILTER, GL_NEAREST );
  glTexParameteri( GL_TEXTURE_2D, GL_TEXTURE_MAG_FILTER, GL_NEAREST );
  glTexStorage2D(GL_TEXTURE_2D, 1, GL_RGBA32F, textureWidth, textureHeight);
  glBindImageTexture(binding, textureIndex, 0, GL_FALSE, 0, GL_WRITE_ONLY, GL_RGBA32F);

  return textureIndex;
}

int bindTexture2DRFloat(int binding, int textureWidth, int textureHeight) {
  int dataSize = sizeof(float) * textureWidth * textureHeight;

  GLuint textureIndex;

  glGenTextures(1, &textureIndex);
  glBindTexture(GL_TEXTURE_2D, textureIndex);
  glTexParameteri( GL_TEXTURE_2D, GL_TEXTURE_MIN_FILTER, GL_NEAREST );
  glTexParameteri( GL_TEXTURE_2D, GL_TEXTURE_MAG_FILTER, GL_NEAREST );
  glTexStorage2D(GL_TEXTURE_2D, 1, GL_R32F, textureWidth, textureHeight);
  glBindImageTexture(binding, textureIndex, 0, GL_FALSE, 0, GL_WRITE_ONLY, GL_R32F);

  return textureIndex;
}

int getStorageBuffer(int bufferIndex, void* buffer, int size) {
  glBindBuffer(GL_SHADER_STORAGE_BUFFER, bufferIndex);
  glGetBufferSubData(GL_SHADER_STORAGE_BUFFER, 0, size, buffer);
  return 0;
}

int getTexture2DRGBAFloatBuffer(int textureIndex, void* buffer, int size) {
  glBindTexture(GL_TEXTURE_2D, textureIndex);
  glGetTexImage(GL_TEXTURE_2D, 0, GL_RGBA, GL_FLOAT, buffer);
  return 0;
}

int getTexture2DRFloatBuffer(int textureIndex, void* buffer, int size) {
  glBindTexture(GL_TEXTURE_2D, textureIndex);
  glGetTexImage(GL_TEXTURE_2D, 0, GL_R, GL_FLOAT, buffer);
  return 0;
}

int deleteBuffer(int bufferIndex) {
  glDeleteBuffers(1, &bufferIndex);
}

int deleteTexture(int textureIndex) {
  glDeleteTextures(1, &textureIndex);
}