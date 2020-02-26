#include <assert.h>
#include <fcntl.h>
#include <stdbool.h>
#include <stdio.h>
#include <string.h>
#include <unistd.h>

#include <gbm.h>

#include <EGL/egl.h>
#include <EGL/eglext.h>

#include <GL/gl.h>
#include <GL/glx.h>
#include <GL/glext.h>

/* a dummy compute shader that does nothing */
#define COMPUTE_SHADER_SRC "          \
#version 310 es\n                                                       \
                                                                        \
layout (local_size_x = 1, local_size_y = 1, local_size_z = 1) in;       \
                                                                        \
void main(void) {                                                       \
   /* awesome compute code here */                                      \
}                                                                       \
"

void checkErrors() {
	GLenum e = glGetError();
	if (e != GL_NO_ERROR) {
		//fprintf(stderr, "OpenGL error: %s (%d)\n", gluErrorString(e), e);
        fprintf(stderr, "OpenGL error: (%x)\n", e);
		exit(20);
	}
}

int main() {
    int fd = open ("/dev/dri/renderD128", O_RDWR);
    assert (fd > 0);
    printf("opened dri device %d\n", fd);

    struct gbm_device *gbm = gbm_create_device (fd);
    assert (gbm != NULL);
    printf("opened gbm device %p\n", gbm);

    /* setup EGL from the GBM device */
   EGLDisplay egl_dpy = eglGetPlatformDisplay (EGL_PLATFORM_GBM_MESA, gbm, NULL);
   assert (egl_dpy != NULL);
   printf("opened egl platform display with gbm device %p\n", egl_dpy);

   EGLint major, minor;
   EGLBoolean res = eglInitialize (egl_dpy, &major, &minor);
   assert (res);
   printf("egl version %d.%d\n", major, minor);

   const char *egl_extension_st = eglQueryString (egl_dpy, EGL_EXTENSIONS);
   assert (strstr (egl_extension_st, "EGL_KHR_create_context") != NULL);
   assert (strstr (egl_extension_st, "EGL_KHR_surfaceless_context") != NULL);

   static const EGLint config_attribs[] = {
      EGL_RENDERABLE_TYPE, EGL_OPENGL_ES3_BIT_KHR,
      EGL_NONE
   };
   EGLConfig cfg;
   EGLint count;

   res = eglChooseConfig (egl_dpy, config_attribs, &cfg, 1, &count);
   assert (res);

   res = eglBindAPI (EGL_OPENGL_ES_API);
   assert (res);

   static const EGLint attribs[] = {
      EGL_CONTEXT_CLIENT_VERSION, 3,
      EGL_NONE
   };

   // see that https://community.arm.com/developer/tools-software/graphics/b/blog/posts/get-started-with-compute-shaders

   // OpenGL version 4.3, forward compatible core profile
	/*int gl3attr[] = {
      GLX_CONTEXT_MAJOR_VERSION_ARB, 4,
      GLX_CONTEXT_MINOR_VERSION_ARB, 3,
      GLX_CONTEXT_PROFILE_MASK_ARB, GLX_CONTEXT_CORE_PROFILE_BIT_ARB,
      GLX_CONTEXT_FLAGS_ARB, GLX_CONTEXT_FORWARD_COMPATIBLE_BIT_ARB,
   None
   };
   wglCreateContextAttribs();
   GLXContext d_ctx = glXCreateContextAttribsARB(egl_dpy, cfg, NULL, true, gl3attr);

	if (!d_ctx) {
		fprintf(stderr, "Couldn't create an OpenGL context\n");
		exit(13);
	}*/

   printf("glx context created\n");

   EGLContext core_ctx = eglCreateContext (egl_dpy,
                                           cfg,
                                           EGL_NO_CONTEXT,
                                           attribs);
   assert (core_ctx != EGL_NO_CONTEXT);

   res = eglMakeCurrent (egl_dpy, EGL_NO_SURFACE, EGL_NO_SURFACE, core_ctx);
   assert (res);

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

   const char *shader_source = COMPUTE_SHADER_SRC;

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

   /* dispatch computation */
   glDispatchCompute (10, 10, 10);
   checkErrors();

   printf ("Compute shader dispatched and finished successfully\n");

   /* free stuff */
   glDeleteProgram (shader_program);
   //eglDestroyContext (egl_dpy, core_ctx);
   eglTerminate (egl_dpy);
   gbm_device_destroy (gbm);
   close (fd);
}