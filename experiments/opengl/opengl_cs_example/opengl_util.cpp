// NOTE:  THERE IS NOTHING COMPUTE SHADER SPECIFIC IN THIS FILE
#include "opengl.h"
#include <X11/Xatom.h>

#include <stdio.h>
#include <string.h>
#include <stdlib.h>

Display *d_dpy;
Window d_win;
GLXContext d_ctx;

GLuint genRenderProg(GLuint texHandle) {
    GLuint progHandle = glCreateProgram();
    GLuint vp = glCreateShader(GL_VERTEX_SHADER);
    GLuint fp = glCreateShader(GL_FRAGMENT_SHADER);

	const char *vpSrc[] = {
		"#version 430\n",
		"in vec2 pos;\
		 out vec2 texCoord;\
		 void main() {\
			 texCoord = pos*0.5f + 0.5f;\
			 gl_Position = vec4(pos.x, pos.y, 0.0, 1.0);\
		 }"
	};

	const char *fpSrc[] = {
		"#version 430\n",
		"uniform sampler2D srcTex;\
		 in vec2 texCoord;\
		 out vec4 color;\
		 void main() {\
			 float c = texture(srcTex, texCoord).x;\
			 color = vec4(c, 1.0, 1.0, 1.0);\
		 }"
	};

    glShaderSource(vp, 2, vpSrc, NULL);
    glShaderSource(fp, 2, fpSrc, NULL);

    glCompileShader(vp);
    int rvalue;
    glGetShaderiv(vp, GL_COMPILE_STATUS, &rvalue);
    if (!rvalue) {
        fprintf(stderr, "Error in compiling vp\n");
        exit(30);
    }
    glAttachShader(progHandle, vp);

    glCompileShader(fp);
    glGetShaderiv(fp, GL_COMPILE_STATUS, &rvalue);
    if (!rvalue) {
        fprintf(stderr, "Error in compiling fp\n");
        exit(31);
    }
    glAttachShader(progHandle, fp);

	glBindFragDataLocation(progHandle, 0, "color");
    glLinkProgram(progHandle);

    glGetProgramiv(progHandle, GL_LINK_STATUS, &rvalue);
    if (!rvalue) {
        fprintf(stderr, "Error in linking sp\n");
        exit(32);
    }   
    
	glUseProgram(progHandle);
	glUniform1i(glGetUniformLocation(progHandle, "srcTex"), 0);

	GLuint vertArray;
    glGenVertexArrays(1, &vertArray);
	glBindVertexArray(vertArray);

	GLuint posBuf;
	glGenBuffers(1, &posBuf);
    glBindBuffer(GL_ARRAY_BUFFER, posBuf);
	float data[] = {
		-1.0f, -1.0f,
		-1.0f, 1.0f,
		1.0f, -1.0f,
		1.0f, 1.0f
	};
    glBufferData(GL_ARRAY_BUFFER, sizeof(float)*8, data, GL_STREAM_DRAW);
	GLint posPtr = glGetAttribLocation(progHandle, "pos");
    glVertexAttribPointer(posPtr, 2, GL_FLOAT, GL_FALSE, 0, 0);
    glEnableVertexAttribArray(posPtr);

	checkErrors("Render shaders");
	return progHandle;
}

GLuint genTexture() {
	// We create a single float channel 512^2 texture
	GLuint texHandle;
	glGenTextures(1, &texHandle);

	glActiveTexture(GL_TEXTURE0);
	glBindTexture(GL_TEXTURE_2D, texHandle);
	glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_MIN_FILTER, GL_LINEAR);
	glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_MAG_FILTER, GL_LINEAR);
	glTexImage2D(GL_TEXTURE_2D, 0, GL_R32F, 512, 512, 0, GL_RED, GL_FLOAT, NULL);

	// Because we're also using this tex as an image (in order to write to it),
	// we bind it to an image unit as well
	glBindImageTexture(0, texHandle, 0, GL_FALSE, 0, GL_WRITE_ONLY, GL_R32F);
	checkErrors("Gen texture");	
	return texHandle;
}

void checkErrors(std::string desc) {
	GLenum e = glGetError();
	if (e != GL_NO_ERROR) {
		fprintf(stderr, "OpenGL error in \"%s\": %s (%d)\n", desc.c_str(), gluErrorString(e), e);
		exit(20);
	}
}

void initGL() {
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

void swapBuffers() {
	glXSwapBuffers(d_dpy, d_win);
    checkErrors("Swapping bufs");
}
