TARGET=./opengl-compute

all: Makefile $(TARGET)

$(TARGET): main.c
	gcc main.c -o $(TARGET) `pkg-config --libs --cflags gl glu x11 gbm egl`

run: $(TARGET)
	$(TARGET)

vk: vk.cpp
	gcc vk.cpp -o vk `pkg-config --libs --cflags vulkan`

clean:
	rm -f $(TARGET)

#		-lSDL2 -lEGL -lGLESv2 -lglapi -ldrm_nouveau -lnx \
#sudo apt install libgles2-mesa-dev libgbm-dev
#`pkg-config --libs --cflags glesv2 egl gbm`

egl: egl.c Makefile
	gcc -o egl egl.c -lOpenGL -lGLU -lEGL

egl-es: egl-es.c Makefile
	gcc -o egl-es egl-es.c -lOpenGL -lGLU -lEGL