package main

import (
	"log"
	"runtime"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
)

const (
	width	= 500
	height	= 500
)

func main() {
	runtime.LockOSThread();

	window := initGlfw();
	defer glfw.Terminate();

	program := initOpenGl();

	for !window.ShouldClose() {
		draw(window, program);
	}
}

// initGlfw initializes glfw and returns a Window object that can be used to render graphics.
func initGlfw() *glfw.Window {
	if err := glfw.Init(); err != nil {
		panic(err);
	}

	glfw.WindowHint(glfw.Resizable, glfw.False);
	glfw.WindowHint(glfw.ContextVersionMajor, 4);
	glfw.WindowHint(glfw.ContextVersionMinor, 1);
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile);
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True);

	window, err := glfw.CreateWindow(width, height, "Conway's Game of Life", nil, nil);
	if err != nil {
		panic(err);
	}
	window.MakeContextCurrent();

	return window;
}

// initOpenGL initializes OpenGL and returns an initialized program
func initOpenGL() uint32 {
	if err := gl.Init(); err != nil {
		panic(err);
	}
	version := gl.GoStr(gl.GetString(gl.VERSION));
	prog := gl.CreateProgram();
	gl.LinkProgram(prog);
	return prog;
}

// draw clears anything that's on the screen before drawing new objects
func draw(window *glfw.Window, program uint32) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT);
	gl.UseProgram(prog);

	glfw.PollEvents();
	window.SwapBuffers();
}

