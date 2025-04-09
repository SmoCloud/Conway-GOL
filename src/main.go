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

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
)

const (
	width     = 500
	height    = 500
	rows      = 100
	cols      = 100
	threshold = 0.15
	fps       = 10 // generations per second

	vertexShaderSource = `
        #version 410
        in vec3 vp;
        void main() {
            gl_Position = vec4(vp, 1.0);
        }
    ` + "\x00"

	fragmentShaderSource = `
        #version 410
        out vec4 frag_colour;
        void main() {
            frag_colour = vec4(0.9, 0.4, 0.7, 1);
        }
    ` + "\x00"
)

var (
	square = []float32{
		-0.5, 0.5, 0, // top   X, Y, Z
		-0.5, -0.5, 0, // left  X, Y, Z
		0.5, -0.5, 0, // right X, Y, Z

		0.5, -0.5, 0,
		0.5, 0.5, 0,
		-0.5, 0.5, 0,
	}
)

// BEGIN - The cell struct and all of its methods start here
type cell struct {
	drawable uint32

	alive    bool
	survives bool

	x int
	y int
}

func (c *cell) draw() {
	if !c.alive {
		return
	}
	gl.BindVertexArray(c.drawable)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(square)/3))
}

// checkState determines if the cell will survive to the next generation
func (c *cell) checkState(cells [][]*cell) {
	c.alive = c.survives
	c.survives = c.alive

	liveCount := c.liveNeighbors(cells)
	if c.alive {
		// Any live cell w/ fewer than two live neighbors dies due to underpopulation
		if liveCount < 2 {
			c.survives = false
		}

		// Any live cell with two to three live neighbors survives to the next generation
		if liveCount == 2 || liveCount == 3 {
			c.survives = true
		}

		// Any live cell with more than three live neighbors dies due to overpopulation
		if liveCount > 3 {
			c.survives = false
		}
	} else {
		// Any dead cell surrounded by three live neighbors is born through reproduction (Happy Birthday!)
		if liveCount == 3 {
			c.survives = true
		}
	}
}

// liveNeighbors returns the amount of surrounding cells that are alive
// parallelization will likely be done here
func (c *cell) liveNeighbors(cells [][]*cell) int {
	var liveCount int
	add := func(x, y int) {
		// Check other side of the board if cell is at the edge ("edge" condition)
		if x == len(cells) {
			x = 0
		} else if x == -1 {
			x = len(cells) - 1
		}
		if y == len(cells[x]) {
			y = 0
		} else if y == -1 {
			y = len(cells[x]) - 1
		}

		if cells[x][y].alive {
			liveCount++
		}
	}

	// Checks all surrounding cells for life
	add(c.x-1, c.y)   // check left cell
	add(c.x+1, c.y)   // check right cell
	add(c.x, c.y+1)   // check cell above
	add(c.x, c.y-1)   // check cell below
	add(c.x-1, c.y+1) // check top-left cell
	add(c.x+1, c.y+1) // check top-right cell
	add(c.x-1, c.y-1) // check bottom-left cell
	add(c.x+1, c.y-1) // check bottom-right cell

	return liveCount
}

// END - The cell struct and all of its methods ends here

func main() {
	runtime.LockOSThread()

	window := initGlfw()
	defer glfw.Terminate()

	program := initOpenGL()

	cells := makeCells()
	for !window.ShouldClose() {
		t := time.Now()

		for x := range cells {
			for _, c := range cells[x] {
				c.checkState(cells)
			}
		}

		draw(cells, window, program)

		time.Sleep(time.Second/time.Duration(fps) - time.Since(t))
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

	window, err := glfw.CreateWindow(width, height, "Conway's Game of Life", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	return window
}

// initOpenGL initializes OpenGL and returns an initialized program
func initOpenGL() uint32 {
	if err := gl.Init(); err != nil {
		panic(err)
	}
	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Println("OpenGL version", version)

	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		panic(err)
	}
	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
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
func draw(cells [][]*cell, window *glfw.Window, program uint32) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.UseProgram(program)

	for x := range cells {
		for _, c := range cells[x] {
			c.draw()
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
func makeCells() [][]*cell {
	rand.New(rand.NewSource(time.Now().UnixNano())) // rand.Seed(seed) has been depreciated, this is the new method for seeding the random number generator

	cells := make([][]*cell, rows, cols)
	for x := range rows {
		for y := range cols {
			c := newCell(x, y)

			c.alive = rand.Float64() < threshold
			c.survives = c.alive

			cells[x] = append(cells[x], c)
		}
	}

	return cells
}

func newCell(x, y int) *cell {
	points := make([]float32, len(square))
	copy(points, square)

	for i := range points {
		var position, size float32
		switch i % 3 {
		case 0:
			size = 1.0 / float32(cols)
			position = float32(x) * size
		case 1:
			size = 1.0 / float32(rows)
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

	return &cell{
		drawable: makeVao(points),

		x: x,
		y: y,
	}
}
