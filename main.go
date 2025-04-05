package main

import (
	"log"
	"runtime"
    "strings"
    "fmt"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
)

const (
	width	= 500
	height	= 500
)

var (
	triangle = []float32 {
		0, 0.5, 0,	// top   X, Y, Z
		-0.5, -0.5, 0,	// left  X, Y, Z
		0.5, -0.5, 0,	// right X, Y, Z
	}
)

func main() {
	runtime.LockOSThread();

	window := initGlfw();
	defer glfw.Terminate();

	program := initOpenGL();

    vao := makeVao(triangle);
	for !window.ShouldClose() {
		draw(vao, window, program);
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
	log.Println("OpenGL version", version);

	prog := gl.CreateProgram();
	gl.LinkProgram(prog);
	return prog;
}

// draw clears anything that's on the screen before drawing new objects
func draw(vao uint32, window *glfw.Window, program uint32) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT);
	gl.UseProgram(program);

    gl.BindVertexArray(vao);
    gl.DrawArrays(gl.TRIANGLES, 0, int32(len(triangle) / 3));

	glfw.PollEvents();
	window.SwapBuffers();
}

// makeVao initializes and returns a vertex array from the points provided.
func makeVao(points []float32) uint32 {
	var vbo uint32;
	gl.GenBuffers(1, &vbo);
    gl.BindBuffer(gl.ARRAY_BUFFER, vbo);
    gl.BufferData(gl.ARRAY_BUFFER, 4*len(points), gl.Ptr(points), gl.STATIC_DRAW);

    var vao uint32;
    gl.GenVertexArrays(1, &vao);
    gl.BindVertexArray(vao);
    gl.EnableVertexAttribArray(0);
    gl.BindBuffer(gl.ARRAY_BUFFER, vbo);
    gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil);

    return vao;
}

func compileShader(source string, shaderType uint32) (uint32, error) {
    shader := gl.CreateShader(shaderType);

    csources, free := gl.Strs(source);
    gl.CompileShader(shader, 1, csources, nil);
    free();
    gl.CompileShader(shader);

    var status uint32;
    gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status);
    if status == gl.FLASE {
        var logLength uint32;
        gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength);

        log := strings.Repeat("\x00", int(logLength + 1));
        gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log));

        return 0, fmt.Errorf("failed to compile %v: %v", source, log);
    }

    return shader, nil;
}

