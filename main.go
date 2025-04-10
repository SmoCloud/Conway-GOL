/*
   Followed along with a tutorial on implementing Conway's Game of Life OpenGL in Go at
   https://kylewbanks.com/blog/tutorial-opengl-with-golang-part-1-hello-opengl
   Tutorial did not include any parallelization or file structuring,
   A lot of the OpenGL boilerplate is similar to the texture compression project from last semester where I used OpenGL in C++ to render the compressed image to the screen.
   I'm really enjoying Go.
*/

package main

import (
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"strings"
	"time"
	"sync"

	gcells "github.com/SmoCloud/Conway-GOL/gol_cells"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
)

func main() {
	runtime.LockOSThread()

	window := initGlfw()
	defer glfw.Terminate()

	program := initOpenGL()

	cells := makeCells()
	wg := new(sync.WaitGroup)
	checkCellState := func(x int, wg *sync.WaitGroup) {
		defer wg.Done()
		for _, c := range cells[x] {
			wg.Add(1)
			go c.CheckState(cells)
			wg.Done()
		}
	}
	for !window.ShouldClose() {
		t := time.Now()
		// Could parallelize the CheckState call, may use waitGroups for this since no data need be sent back to main
		for x := range cells {
			wg.Add(1)
			go checkCellState(x, wg)
		}
		wg.Wait()
		draw(cells, window, program)

		time.Sleep(time.Second / time.Duration(gcells.Fps) - time.Since(t))
	}
}

// initGlfw initializes glfw and returns a Window object that can be used to render graphics.
func initGlfw() *glfw.Window {
	if err := glfw.Init(); err != nil {
		panic(err)
	}

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(gcells.Width, gcells.Height, "Conway's Game of Life", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()
	glfw.SwapInterval(glfw.True)

	return window
}

// initOpenGL initializes OpenGL and returns an initialized program
func initOpenGL() uint32 {
	if err := gl.Init(); err != nil {
		panic(err)
	}
	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Println("OpenGL version", version)

	vertexShader, err := compileShader(gcells.VertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		panic(err)
	}
	fragmentShader, err := compileShader(gcells.FragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		panic(err)
	}

	prog := gl.CreateProgram()
	gl.AttachShader(prog, vertexShader)
	gl.AttachShader(prog, fragmentShader)
	gl.LinkProgram(prog)
	return prog
}

// draw clears anything that's on the screen before drawing new objects
// Cannot parallelize draws as OpenGL requires operations to happen on a single thread
func draw(cells [][]*gcells.Cell, window *glfw.Window, program uint32) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.UseProgram(program)

	for x := range cells {
		for _, c := range cells[x] {
			c.Draw()
		}
	}

	glfw.PollEvents()
	window.SwapBuffers()
}

// makeVao initializes and returns a vertex array from the points provided.
func makeVao(points []float32) uint32 {
	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(points), gl.Ptr(points), gl.STATIC_DRAW)

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	gl.EnableVertexAttribArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil)

	return vao
}

// compileShader will send the shader source code to the GPU for compilation on the GPU (shaders handle vertex points of drawn objects as well as their color)
func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}

// this function will create the game board grid
// Can parallelize here (use WaitGroup instead of a channel as there's no need to pass data back to main, it already gets stored in slice reference)
func makeCells() [][]*gcells.Cell {
	rand.New(rand.NewSource(time.Now().UnixNano())) // rand.Seed(seed) has been depreciated, this is the new method for seeding the random number generator
	cells := make([][]*gcells.Cell, gcells.Cols)
	for x := range gcells.Cols {
		for y := range gcells.Rows {
			// go makeCellsHelperHelper(cells, x, y, wg)
			c := newCell(x, y)
	
			c.Alive = rand.Float64() < gcells.Threshold
			c.Survives = c.Alive
	
			cells[x] = append(cells[x], c)
		}
	}

	return cells
}

// func makeCellsHelper(cells [][]*gcells.Cell, x int) {
	
// }

// func makeCellsHelperHelper(cells [][]*gcells.Cell, x, y int, wg* sync.WaitGroup) {
// 	defer wg.Done()
	
// }

// This function creates a new cell
func newCell(x, y int) *gcells.Cell {
	points := make([]float32, len(gcells.Square))
	copy(points, gcells.Square)

	for i := range points {
		var position, size float32
		switch i % 3 {
		case 0:
			size = 1.0 / float32(gcells.Cols)
			position = float32(x) * size
		case 1:
			size = 1.0 / float32(gcells.Rows)
			position = float32(y) * size
		default:
			continue
		}

		if points[i] < 0 {
			points[i] = (position * 2) - 1
		} else {
			points[i] = ((position + size) * 2) - 1
		}
	}

	return &gcells.Cell{
		Drawable: makeVao(points),

		X: x,
		Y: y,
	}
}
