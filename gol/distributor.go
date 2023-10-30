package gol

import (
	"fmt"
	"math"
	"time"

	//"strconv"

	"uk.ac.bris.cs/gameoflife/util"
)

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- uint8
	ioInput    <-chan uint8
}

func advanceWorld(world [][]byte, p Params, startY, endY int) [][]byte {

	//fmt.Printf("sY: %d, eY: %d\n", startY, endY)

	//world the size of this routines parameters
	newworld := make([][]byte, endY-startY)
	for i := range newworld {
		newworld[i] = make([]byte, p.ImageHeight)
	}

	// check each cell in range
	for i := startY; i < endY; i++ {
		for j := range world[i] {

			sum := 0

			// alive neighbour count
			for m := -1; m <= 1; m++ {
				for n := -1; n <= 1; n++ {
					if m != 0 || n != 0 {
						dx := (i + m + p.ImageWidth) % (p.ImageWidth)
						dy := (j + n + p.ImageHeight) % (p.ImageHeight)

						//fmt.Printf("%d-%d | i: %d, i-y: %d: j: %d, dx: %d, dy: %d\n", startY, endY, i, i-startY, j, dx, dy)

						sum += (int(world[dx][dy]) / 255)

					}
				}
			}

			//apply rules
			if sum == 3 || (int(world[i][j])/255)+sum == 3 {
				// correct for i/j being offset
				//fmt.Printf("%d-%d | i: %d, i-y: %d: j: %d, j-y:%d\n", startY, endY, i, i-startY, j, j-endY)
				ni := i - startY
				newworld[ni][j] = 255
			}
		}
	}

	return newworld

}

func worker(world [][]byte, p Params, startY, endY int, out chan<- [][]byte) {
	output := advanceWorld(world, p, startY, endY)
	out <- output
}

func getAliveCells(world [][]byte) []util.Cell {

	cells := []util.Cell{}

	for i := range world {
		for j := range world[i] {
			if world[i][j] == 255 {
				cells = append(cells, util.Cell{X: j, Y: i})
			}
		}
	}
	return cells
}

func aliveTicker(out chan<- bool) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			out <- true
		}
	}
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {

	c.ioCommand <- ioInput
	c.ioFilename <- fmt.Sprintf("%dx%d", p.ImageWidth, p.ImageHeight)

	// make 2d slice to hold world
	world := make([][]byte, p.ImageWidth)
	// initialise world
	for i := range world {
		world[i] = make([]byte, p.ImageHeight)
		for j := range world[i] {
			world[i][j] = <-c.ioInput
		}
	}

	// to track turns
	turn := 0

	// to track whether alivecells should be counted
	count := make(chan bool)
	go aliveTicker(count)

	// only use a power of 2 for thread count
	// we cannot divide a 16x16 image by 3 for example...
	bt := int(math.Log2(float64(p.Threads)))
	vthreads := int(math.Pow(2, float64(bt)))

	// TODO: Execute all turns of the Game of Life.
	for turn = 0; turn < p.Turns; turn++ {

		// temp world to hold new data
		var tworld [][]byte

		// switch on thread count
		if vthreads == 1 {
			tworld = advanceWorld(world, p, 0, p.ImageHeight)
		} else {
			chans := make([]chan [][]byte, vthreads)
			x := p.ImageHeight / vthreads

			// allocate each thread a strip of the world to work on
			for i := 0; i < vthreads; i++ {
				chans[i] = make(chan [][]byte)
				go worker(world, p, (x * i), (x*i + x), chans[i])
			}
			for c := 0; c < vthreads; c++ {
				part := <-chans[c]
				tworld = append(tworld, part...)
			}
		}

		world = tworld

		select {
		case <-count:
			c.events <- AliveCellsCount{turn + 1, len(getAliveCells(world))}
		default:
		}

		c.events <- TurnComplete{CompletedTurns: turn + 1}
	}

	// TODO: Report the final state using FinalTurnCompleteEvent.
	c.events <- FinalTurnComplete{turn, getAliveCells(world)}

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{turn, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}
