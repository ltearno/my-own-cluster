#define GL_GLEXT_PROTOTYPES
#include <fcntl.h>
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/stat.h>

#include <gbm.h>

#include <EGL/egl.h>
#include <EGL/eglext.h>

#include <GL/gl.h>
#include <GL/glext.h>

typedef struct {
  int renderDeviceFd;
  struct gbm_device *gbm;
   EGLDisplay egl_dpy;
   EGLConfig config;
} DeviceInformation;

typedef struct {
   EGLContext core_ctx;
} ContextInformation;

int initOpenGLDevice(DeviceInformation *info);
int initOpenGLContext(DeviceInformation *dInfo, ContextInformation *info);
int setContext(DeviceInformation *dInfo, ContextInformation *info);
int clearContext(DeviceInformation *dInfo, ContextInformation *info);
int compileAndBindShader(const char *shader_source);

int bindStorageBuffer(int binding, void* buffer, int size);
int bindTexture2DRGBAFloat(int binding, int textureWidth, int textureHeight);
int bindTexture2DRFloat(int binding, int textureWidth, int textureHeight);
int getStorageBuffer(int bufferIndex, void* buffer, int size);
int getTexture2DRGBAFloatBuffer(int textureIndex, void* buffer, int size);
int getTexture2DRFloatBuffer(int textureIndex, void* buffer, int size);
int deleteBuffer(int bufferIndex);
int deleteTexture(int textureIndex);