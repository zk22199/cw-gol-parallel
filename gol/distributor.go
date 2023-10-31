package gol

import (
	"fmt"
	"time"

	"uk.ac.bris.cs/gameoflife/util"
)

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- uint8
	ioInput    <-chan uint8
	keyPress   <-chan rune
}

func distribute(world [][]byte, p Params, c distributorChannels, t int) [][]byte {

	// initialise slice of channels to maintain order
	// when sending tasks to worker threads
	channels := make([]chan [][]byte, p.Threads)
	for i := range channels {
		channels[i] = make(chan [][]byte)
	}

	// this is a rough even split to separate between workers
	heightDiff := p.ImageHeight / p.Threads

	// sets up workers for all except last slice
	for i := 0; i < p.Threads-1; i++ {
		go worker(world, p, c, t, i*heightDiff, (i+1)*heightDiff, channels[i])
	}

	// sets up worker for last slice, necessary to correct
	// for inconsistencies with rounding
	go worker(world, p, c, t, (p.Threads-1)*heightDiff, p.ImageHeight, channels[p.Threads-1])

	var newWorld [][]byte

	// appends each individual slice to the resulting next state
	// maintains order
	for i := 0; i < p.Threads; i++ {
		thisSlice := <-channels[i]
		newWorld = append(newWorld, thisSlice...)
	}
	return newWorld
}

func getAliveCells(world [][]byte) []util.Cell {

	cells := []util.Cell{}

	// counts the number of cells which correspond
	// to an on value (255)
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
	// ticker that sends a true signal down an output channel every 2 seconds
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			out <- true
		}
	}
}

func saveBoard(world [][]byte, p Params, c distributorChannels) {

}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {

	c.ioCommand <- ioInput
	c.ioFilename <- fmt.Sprintf("%d%s%d", p.ImageWidth, "x", p.ImageHeight)

	// make 2d slice to hold world
	// pipe input for each value from the stream input channel
	world := make([][]byte, p.ImageHeight)
	for i := range world {
		world[i] = make([]byte, p.ImageWidth)
		for j := range world[i] {
			world[i][j] = <-c.ioInput
		}
	}

	turn := 0

	// flip all alive cells
	for i := range world {
		for j := range world[i] {
			if world[i][j] == 255 {
				c.events <- CellFlipped{CompletedTurns: turn, Cell: util.Cell{X: j, Y: i}}
			}
		}
	}

	// to track whether alivecells should be counted
	count := make(chan bool)
	go aliveTicker(count)

	// distributes tasks for each turn depending on number of threads
	for turn = 0; turn < p.Turns; turn++ {
		world = distribute(world, p, c, turn)

		c.events <- TurnComplete{CompletedTurns: turn}

		// selects appropriate action based on keyboard presses/ ticker
		select {
		case <-count: //ticker call
			c.events <- AliveCellsCount{turn + 1, len(getAliveCells(world))}
		case key := <-c.keyPress:
			switch key {
			case 's':
				saveBoard(world, p, c)
			case 'q':
				saveBoard(world, p, c)
			case 'p':
				time.Sleep(2)
			}
		default:
		}

	}

	// Report the final turn being complete
	c.events <- FinalTurnComplete{turn, getAliveCells(world)}

	// save the world as a pgm file
	saveBoard(world, p, c)

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{turn, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}
