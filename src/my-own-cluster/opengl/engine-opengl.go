package opengl

/*
#cgo LDFLAGS: -lOpenGL -lGLU -lgbm -lEGL
#define GL_GLEXT_PROTOTYPES
#include <fcntl.h>
#include <EGL/egl.h>
#include <EGL/eglext.h>
#include <GL/gl.h>
#include <sys/stat.h>
#include <stdlib.h>

#include "glext.h"

#include <gbm.h>

#include <stdio.h>

void err(int n, char* s){
	printf("err %d %s\n", n, s);
}

int checkErrors(const char* when) {
	GLenum e = glGetError();
	if (e != GL_NO_ERROR) {
		//fprintf(stderr, "OpenGL error: %s (%d)\n", gluErrorString(e), e);
        fprintf(stderr, "OpenGL error when %s: (0x%x)\n", when, e);
		return -1;
  }

  return 0;
}

  static const EGLint configAttribs[] = {
          //EGL_RENDERABLE_TYPE, EGL_OPENGL_ES3_BIT_KHR,
          EGL_NONE
  };

  static const EGLint pbufferAttribs[] = {
        EGL_WIDTH, 9,
        EGL_HEIGHT, 9,
        EGL_NONE,
  };


int initGLByEGL() {
  // 1. Initialize EGL
  //unsetenv("DISPLAY");
  EGLDisplay eglDpy = eglGetDisplay(EGL_DEFAULT_DISPLAY);
  EGLint major, minor;
  EGLBoolean res = eglInitialize(eglDpy, &major, &minor);
  if(major!=1) {
    printf("trying initialize with gbm and dri\n");
    int fd = open ("/dev/dri/renderD128", O_RDWR);
    if (fd < 0) {
      printf("cannot open /dev/dri/renderD128 device\n");
      return -3;
    }
    printf("opened dri device %d\n", fd);

    struct gbm_device *gbm = gbm_create_device (fd);
    if (gbm == NULL) {
      printf("cannot create gbm device\n");
      return -4;
    }
    printf("opened gbm device %p\n", gbm);

    eglDpy = eglGetPlatformDisplay (EGL_PLATFORM_GBM_MESA, gbm, NULL);

    EGLBoolean res = eglInitialize(eglDpy, &major, &minor);
    if(major!=1) {
      return -11;
    }
  }
  printf("egl version %d.%d\n", major, minor);



  printf("EGL_CLIENT_APIS: '%s'\n", eglQueryString(eglDpy, EGL_CLIENT_APIS));
  printf("EGL_EXTENSIONS: '%s'\n", eglQueryString(eglDpy, EGL_EXTENSIONS));
  printf("EGL_VENDOR: '%s'\n", eglQueryString(eglDpy, EGL_VENDOR));
  printf("EGL_VERSION: '%s'\n", eglQueryString(eglDpy, EGL_VERSION));

  // 2. Select an appropriate configuration
  EGLint configCount;
  eglChooseConfig (eglDpy, configAttribs, NULL, 0, &configCount);
  printf("config_count: %d\n", configCount);

  EGLConfig *eglConfigs = (EGLConfig*) malloc(sizeof(EGLConfig) * configCount);
  eglChooseConfig (eglDpy, configAttribs, eglConfigs, configCount, &configCount);

  EGLConfig eglCfg = eglConfigs[0];
  //eglChooseConfig(eglDpy, configAttribs, &eglCfg, 1, &configCount);

  // 3. Create a surface (used in eglMakeCurrent but not required since we only work with SSBO for now...)
  //EGLSurface eglSurf = eglCreatePbufferSurface(eglDpy, eglCfg, pbufferAttribs);

  // 4. Bind the API
  eglBindAPI(EGL_OPENGL_API);

  // 5. Create a context and make it current
  EGLint eglAttrs[] = {
    //EGL_CONTEXT_CLIENT_VERSION, 3,
    EGL_NONE
  };

  EGLContext eglCtx = eglCreateContext(eglDpy, eglCfg, EGL_NO_CONTEXT, eglAttrs);

  eglMakeCurrent(eglDpy, EGL_NO_SURFACE, EGL_NO_SURFACE, eglCtx);

  // TODO should do that at the end
  //eglTerminate(eglDpy);

  return 0;
}

int runShader(const char *inputData, int inputDataLen, char *outputData, int outputDataLen, const char *shader_source, int dispatchSizeX, int dispatchSizeY, int dispatchSizeZ)
{
  printf("GL_VERSION: '%s'\n", glGetString(GL_VERSION));
  printf("GL_VENDOR: '%s'\n", glGetString(GL_VENDOR));
  printf("GL_RENDERER: '%s'\n", glGetString(GL_RENDERER));
  printf("GL_EXTENSIONS: '%s'\n", glGetString(GL_EXTENSIONS));

  GLuint inIndex;
  glGenBuffers(1, &inIndex);
  if(checkErrors("gen buffers")!=0) return -1;

  GLuint outIndex;
  glGenBuffers(1, &outIndex);
  if(checkErrors("gen buffers")!=0) return -1;

  printf("buffers: %d %d\n", inIndex, outIndex);

  glBindBuffer(GL_SHADER_STORAGE_BUFFER, inIndex);
  if(checkErrors("bind buffer")!=0) return -1;
  glBufferData(GL_SHADER_STORAGE_BUFFER, inputDataLen, inputData, GL_DYNAMIC_COPY);
  if(checkErrors("buffer data")!=0) return -1;
  glBindBufferBase(GL_SHADER_STORAGE_BUFFER, 0, inIndex);
  if(checkErrors("bind buffer base")!=0) return -1;

  glBindBuffer(GL_SHADER_STORAGE_BUFFER, outIndex);
  if(checkErrors("bind buffer")!=0) return -1;
  glBufferData(GL_SHADER_STORAGE_BUFFER, outputDataLen, NULL, GL_DYNAMIC_COPY);
  glBindBufferBase(GL_SHADER_STORAGE_BUFFER, 1, outIndex);
  if(checkErrors("creating and binding buffers")!=0) return -1;

  // setup a compute shader
  GLuint compute_shader = glCreateShader (GL_COMPUTE_SHADER);
  if(checkErrors("creating shader")!=0) return -1;

  glShaderSource (compute_shader, 1, &shader_source, NULL);
  if(checkErrors("giving shader source")!=0) return -1;

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
    return -1;
  }
  if(checkErrors("compiling shader")!=0) return -1;

  GLuint shader_program = glCreateProgram ();

  glAttachShader (shader_program, compute_shader);
  if(checkErrors("attaching shader")!=0) return -1;

  glLinkProgram (shader_program);
  if(checkErrors("linking shader")!=0) return -1;

  glDeleteShader (compute_shader);

  glUseProgram (shader_program);
  if(checkErrors("using shader in program")!=0) return -1;

  // dispatch computation
  glDispatchCompute (dispatchSizeX, dispatchSizeY, dispatchSizeZ);
  if(checkErrors("dispatching compute")!=0) return -1;

  glBindBuffer(GL_SHADER_STORAGE_BUFFER, outIndex);
  if(checkErrors("binding buffer")!=0) return -1;
  glGetBufferSubData(GL_SHADER_STORAGE_BUFFER, 0, outputDataLen, outputData);
  if(checkErrors("getting buffer sub data")!=0) return -1;
  return 0;
}
*/
import "C"
import (
	"errors"
	"my-own-cluster/common"
	"unsafe"
)

type GLSLOpenGLProcessContext struct {
	Fctx *common.FunctionExecutionContext
}

type GLSLOpenGLEngine struct {
}

// TODO manage EGL contexts and free resources when used !! This is highly not done here (dirty crap code)

func NewGLSLOpenGLEngine() (*GLSLOpenGLEngine, error) {
	if C.initGLByEGL() != 0 {
		return nil, errors.New("cannot instantiate OpenGL")
	}

	return &GLSLOpenGLEngine{}, nil
}

func (e *GLSLOpenGLEngine) PrepareContext(fctx *common.FunctionExecutionContext) (common.ExecutionEngineContext, error) {
	return &GLSLOpenGLProcessContext{
		Fctx: fctx,
	}, nil
}

func (c *GLSLOpenGLProcessContext) Run() error {
	C.initGLByEGL()

	inputBuffer := c.Fctx.Orchestrator.GetExchangeBuffer(c.Fctx.InputExchangeBufferID).GetBuffer()

	//fmt.Println(inputBuffer)
	//floats := (*[1 << 30]float32)(unsafe.Pointer(&inputBuffer[0]))[:1024:1024]
	/*for i := 0; i < 10; i++ {
		//fmt.Printf("%f\n", floats[i])

		bits := binary.LittleEndian.Uint32(inputBuffer[i*4 : (i+1)*4])
		fl := math.Float32frombits(bits)
		//fmt.Println(fl)
	}*/

	// by default we use input buffer size, but that should be changed
	outputData := make([]byte, len(inputBuffer))
	//fmt.Printf("output length : %d\n", len(outputData))

	C.runShader(
		(*C.char)(unsafe.Pointer(&inputBuffer[0])), C.int(len(inputBuffer)),
		(*C.char)(unsafe.Pointer(&outputData[0])), C.int(len(outputData)),
		C.CString(string(c.Fctx.CodeBytes)),
		C.int(len(inputBuffer)/4 /*float size, hardcoded*/), C.int(1), C.int(1))

	//fmt.Println(outputData)
	/*for i := 0; i < 10; i++ {
		bits := binary.LittleEndian.Uint32(outputData[i*4 : (i+1)*4])
		fl := math.Float32frombits(bits)
		//fmt.Println(fl)
	}*/

	c.Fctx.Orchestrator.GetExchangeBuffer(c.Fctx.OutputExchangeBufferID).Write(outputData)
	return nil
}
