#ifndef _OPENGL_H
#define _OPENGL_H

#define GL_GLEXT_PROTOTYPES
#define GLX_GLXEXT_PROTOTYPES

#include <X11/Xlib.h>
#include <GL/gl.h>
#include <GL/glx.h>
#include <GL/glu.h>

// This includes the new stuff, supplied by the application
#include "GL/glext.h"

#include <string>

#define WIN_WIDTH 800
#define WIN_HEIGHT 600

void initGL();
void swapBuffers();

// Return handles
GLuint genTexture();
GLuint genRenderProg(GLuint); // Texture as the param
GLuint genComputeProg(GLuint);

void checkErrors(std::string);

#endif
