package cells

import (
	// "runtime"

	"github.com/go-gl/gl/v4.1-core/gl"
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