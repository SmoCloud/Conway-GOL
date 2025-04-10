package gcells

import (
	"github.com/go-gl/gl/v4.1-core/gl"
)

const (
	Threshold = 0.15
	Fps       = 60 // generations per second

	VertexShaderSource = `
        #version 410
        in vec3 vp;
        void main() {
            gl_Position = vec4(vp, 1.0);
        }
    ` + "\x00"

	FragmentShaderSource = `
        #version 410
        out vec4 frag_colour;
        void main() {
            frag_colour = vec4(0.9, 0.4, 0.7, 1);
        }
    ` + "\x00"
)

var (
	Square = []float32{
		-0.5, 0.5, 0, // top   X, Y, Z
		-0.5, -0.5, 0, // left  X, Y, Z
		0.5, -0.5, 0, // right X, Y, Z

		0.5, -0.5, 0,
		0.5, 0.5, 0,
		-0.5, 0.5, 0,
	}
	Width  = 500
	Height = 500
	Rows   = 100
	Cols   = 100
)

// BEGIN - The cell struct and all of its methods start here
type Cell struct {
	Drawable uint32

	Alive    bool
	Survives bool

	X int
	Y int
}

func (c *Cell) Draw() {
	if !c.Alive {
		return
	}
	gl.BindVertexArray(c.Drawable)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(Square)/3))
}

// checkState determines if the cell will survive to the next generation
func (c *Cell) CheckState(cells [][]*Cell) {
	c.Alive = c.Survives
	c.Survives = c.Alive

	liveCount := c.LiveNeighbors(cells)
	if c.Alive {
		// Any live cell w/ fewer than two live neighbors dies due to underpopulation
		if liveCount < 2 {
			c.Survives = false
		}

		// Any live cell with two to three live neighbors survives to the next generation
		if liveCount == 2 || liveCount == 3 {
			c.Survives = true
		}

		// Any live cell with more than three live neighbors dies due to overpopulation
		if liveCount > 3 {
			c.Survives = false
		}
	} else {
		// Any dead cell surrounded by three live neighbors is born through reproduction (Happy Birthday!)
		if liveCount == 3 {
			c.Survives = true
		}
	}
}

func (c *Cell) countAlive(x, y int, cells [][]*Cell, liveCount chan<- int) {
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

	if cells[x][y].Alive {
		liveCount <- 1
	} else {
		liveCount <- 0
	}
}

// liveNeighbors returns the amount of surrounding cells that are alive
// parallelization will likely be done here
func (c *Cell) LiveNeighbors(cells [][]*Cell) int {
	liveChan := make(chan int, 8)
	living := 0

	// Checks all surrounding cells for life
	go c.countAlive(c.X-1, c.Y, cells, liveChan)   // check left cell
	go c.countAlive(c.X+1, c.Y, cells, liveChan)   // check right cell
	go c.countAlive(c.X, c.Y+1, cells, liveChan)   // check cell above
	go c.countAlive(c.X, c.Y-1, cells, liveChan)   // check cell below
	go c.countAlive(c.X-1, c.Y+1, cells, liveChan) // check top-left cell
	go c.countAlive(c.X+1, c.Y+1, cells, liveChan) // check top-right cell
	go c.countAlive(c.X-1, c.Y-1, cells, liveChan) // check bottom-left cell
	go c.countAlive(c.X+1, c.Y-1, cells, liveChan) // check bottom-right cell
	for range 8 {
		living += <-liveChan
	}

	return living
}
