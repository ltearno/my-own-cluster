TARGET=./opengl-compute

all: Makefile $(TARGET)

$(TARGET): main.c
	gcc main.c -o $(TARGET) `pkg-config --libs --cflags gbm egl gl glu`

run: $(TARGET)
	$(TARGET)

clean:
	rm -f $(TARGET)

#		-lSDL2 -lEGL -lGLESv2 -lglapi -ldrm_nouveau -lnx \
#sudo apt install libgles2-mesa-dev libgbm-dev
#`pkg-config --libs --cflags glesv2 egl gbm`