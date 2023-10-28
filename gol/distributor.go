package gol

import (
	"fmt"
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

func executeTurn(world [][]byte, p Params) [][]byte {
	newworld := make([][]byte, p.ImageWidth)
	for i := range newworld {
		newworld[i] = make([]byte, p.ImageHeight)
	}

	// check each cell in world
	for i := range world {
		for j := range world[i] {

			sum := 0

			// alive neighbour count
			for m := -1; m <= 1; m++ {
				for n := -1; n <= 1; n++ {
					if m != 0 || n != 0 {
						dx := (i + m + p.ImageWidth) % p.ImageWidth
						dy := (j + n + p.ImageHeight) % p.ImageHeight

						sum += (int(world[dx][dy]) / 255)

					}
				}
			}

			//apply rules
			if sum == 3 || (int(world[i][j])/255)+sum == 3 {
				newworld[i][j] = 255
			}
		}
	}

	return newworld
}

func getAliveCells(world [][]byte, p Params) []util.Cell {

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

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {

	c.ioCommand <- ioInput
	c.ioFilename <- fmt.Sprintf("%d%s%d", p.ImageWidth, "x", p.ImageHeight)

	// make 2d slice to hold world
	world := make([][]byte, p.ImageWidth)
	for i := range world {
		world[i] = make([]byte, p.ImageHeight)
		for j := range world[i] {
			world[i][j] = <-c.ioInput
		}
	}

	turn := 0

	// TODO: Execute all turns of the Game of Life.
	for turn = 0; turn < p.Turns; turn++ {
		world = executeTurn(world, p)
		c.events <- TurnComplete{CompletedTurns: turn}
	}

	// TODO: Report the final state using FinalTurnCompleteEvent.
	c.events <- FinalTurnComplete{turn, getAliveCells(world, p)}

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{turn, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}
