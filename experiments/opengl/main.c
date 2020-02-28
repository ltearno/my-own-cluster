#include <assert.h>
#include <fcntl.h>
#include <stdbool.h>
#include <stdio.h>
#include <string.h>
#include <unistd.h>
#include <sys/stat.h>

#include <X11/Xlib.h>
#include <X11/Xatom.h>

#include <gbm.h>

#include <EGL/egl.h>
#include <EGL/eglext.h>

#include <GL/gl.h>
#include <GL/glx.h>
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

void initGL();

int main() {
   //initGL();
   //exit(0);

   EGLDisplay egl_dpy;

   if(false) {
      int fd = open ("/dev/dri/renderD128", O_RDWR);
      assert (fd > 0);
      printf("opened dri device %d\n", fd);

      struct gbm_device *gbm = gbm_create_device (fd);
      assert (gbm != NULL);
      printf("opened gbm device %p\n", gbm);

      /* setup EGL from the GBM device */
      egl_dpy = eglGetPlatformDisplay (EGL_PLATFORM_GBM_MESA, gbm, NULL);
   }
   else {
      Display* x_dpy = XOpenDisplay(NULL);
      egl_dpy = eglGetPlatformDisplay(EGL_PLATFORM_X11_KHR, x_dpy, NULL);
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

   static const EGLint config_attribs[] = {
      EGL_RENDERABLE_TYPE, EGL_OPENGL_ES_BIT, //EGL_OPENGL_ES3_BIT_KHR,
      EGL_NONE
   };
   
   EGLint configCount;
   res = eglChooseConfig (egl_dpy, config_attribs, NULL, 0, &configCount);
   printf("config_count: %d\n", configCount);

   EGLConfig *configs = (EGLConfig*) malloc(sizeof(EGLConfig) * configCount);

   res = eglChooseConfig (egl_dpy, config_attribs, configs, configCount, &configCount);
   assert (res);

   for(int i=0; i<configCount; i++) {
      printf("config %d\n", i);
      EGLint configValue;
      eglGetConfigAttrib(egl_dpy, configs[i], EGL_RENDERABLE_TYPE, &configValue);
      printf("EGL_RENDERABLE_TYPE: %d\n", configValue);
   }

   // EGL_OPENGL_API, EGL_OPENGL_ES_API, or EGL_OPENVG_API
   res = eglBindAPI (EGL_OPENGL_API);
   assert (res);

   static const EGLint attribs[] = {
      //EGL_CONTEXT_CLIENT_VERSION, 3,
      EGL_NONE
   };

   EGLContext core_ctx = eglCreateContext (egl_dpy,
                                           configs[0],
                                           EGL_NO_CONTEXT,
                                           attribs);
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




/* prepare input and output */
   const int dataSize = 1;
   float *in1 = malloc(sizeof(float) * dataSize);
   float *in2 = malloc(sizeof(float) * dataSize);
   for(int i=0;i<dataSize; i++ ){
      in1[i] = 0;
      in2[i] = 1;
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

   /* dispatch computation */
   glDispatchCompute (dataSize, 1, 1);
   checkErrors();

   glMemoryBarrier(GL_BUFFER_UPDATE_BARRIER_BIT);
   checkErrors();

   printf ("Compute shader dispatched and finished successfully\n");

   glBindBuffer(GL_SHADER_STORAGE_BUFFER, outIndex);
   checkErrors();
   float* tmp = malloc(sizeof(float) * dataSize);
   glGetBufferSubData(GL_SHADER_STORAGE_BUFFER, 0, sizeof(float) * dataSize, tmp);
   printf("tmp buffer: %p\n", tmp);
   for(int i=0;i<10;i++)
      printf("tmp[%d] = %f\n", i, tmp[i]);

   float *outBound = (float*) glMapBuffer(GL_SHADER_STORAGE_BUFFER, GL_READ_ONLY);
   checkErrors();
   printf("outBound buffer: %p\n", outBound);
   for(int i=8;i<10;i++)
      printf("outBound[%d] = %f\n", i, outBound[i]);
   glUnmapBuffer(GL_SHADER_STORAGE_BUFFER);

   /* free stuff */
   glDeleteProgram (shader_program);
   //eglDestroyContext (egl_dpy, core_ctx);
   eglTerminate (egl_dpy);
   //gbm_device_destroy (gbm);
   //close (fd);
}




void initGL() {
   Display *d_dpy;
   Window d_win;
   GLXContext d_ctx;

	if (!(d_dpy = XOpenDisplay(NULL))) {
		fprintf(stderr, "Couldn't open X11 display\n");
		exit(10);
	}

	int attr[] = {
		GLX_RGBA,
		GLX_RED_SIZE, 1,
		GLX_GREEN_SIZE, 1,
		GLX_BLUE_SIZE, 1,
		GLX_DOUBLEBUFFER,
		None
	};

	int scrnum = DefaultScreen(d_dpy);
	Window root = RootWindow(d_dpy, scrnum);
    
	int elemc;
	GLXFBConfig *fbcfg = glXChooseFBConfig(d_dpy, scrnum, NULL, &elemc);
	if (!fbcfg) {
		fprintf(stderr, "Couldn't get FB configs\n");
		exit(11);
	}

	XVisualInfo *visinfo = glXChooseVisual(d_dpy, scrnum, attr);

	if (!visinfo) {
		fprintf(stderr, "Couldn't get a visual\n");
		exit(12);
	}

	// Window parameters
	XSetWindowAttributes winattr;
	winattr.background_pixel = 0;
	winattr.border_pixel = 0;
	winattr.colormap = XCreateColormap(d_dpy, root, visinfo->visual, AllocNone);
	winattr.event_mask = StructureNotifyMask | ExposureMask | KeyPressMask;
	unsigned long mask = CWBackPixel | CWBorderPixel | CWColormap | CWEventMask;

   const int WIN_WIDTH = 800;
   const int WIN_HEIGHT = 600;
	printf("Window depth %d, %dx%d\n", visinfo->depth, WIN_WIDTH, WIN_HEIGHT);
	d_win = XCreateWindow(d_dpy, root, -1, -1, WIN_WIDTH, WIN_HEIGHT, 0, 
			visinfo->depth, InputOutput, visinfo->visual, mask, &winattr);

	// OpenGL version 4.3, forward compatible core profile
	int gl3attr[] = {
        GLX_CONTEXT_MAJOR_VERSION_ARB, 4,
        GLX_CONTEXT_MINOR_VERSION_ARB, 3,
        GLX_CONTEXT_PROFILE_MASK_ARB, GLX_CONTEXT_CORE_PROFILE_BIT_ARB,
        GLX_CONTEXT_FLAGS_ARB, GLX_CONTEXT_FORWARD_COMPATIBLE_BIT_ARB,
		None
    };

	d_ctx = glXCreateContextAttribsARB(d_dpy, fbcfg[0], NULL, true, gl3attr);

	if (!d_ctx) {
		fprintf(stderr, "Couldn't create an OpenGL context\n");
		exit(13);
	}

	XFree(visinfo);

	// Setting the window name
	XTextProperty windowName;
	windowName.value = (unsigned char *) "OpenGL compute shader demo";
	windowName.encoding = XA_STRING;
	windowName.format = 8;
	windowName.nitems = strlen((char *) windowName.value);

	XSetWMName(d_dpy, d_win, &windowName);

	XMapWindow(d_dpy, d_win);
	glXMakeCurrent(d_dpy, d_win, d_ctx);
	
	printf("OpenGL:\n\tvendor %s\n\trenderer %s\n\tversion %s\n\tshader language %s\n",
			glGetString(GL_VENDOR), glGetString(GL_RENDERER), glGetString(GL_VERSION),
			glGetString(GL_SHADING_LANGUAGE_VERSION));

	// Finding the compute shader extension
	int extCount;
	glGetIntegerv(GL_NUM_EXTENSIONS, &extCount);
	bool found = false;
	for (int i = 0; i < extCount; ++i)
		if (!strcmp((const char*)glGetStringi(GL_EXTENSIONS, i), "GL_ARB_compute_shader")) {
			printf("Extension \"GL_ARB_compute_shader\" found\n");
			found = true;
			break;
		}

	if (!found) {
		fprintf(stderr, "Extension \"GL_ARB_compute_shader\" not found\n");
		exit(14);
	}

	glViewport(0, 0, WIN_WIDTH, WIN_HEIGHT);

	checkErrors("Window init");
}