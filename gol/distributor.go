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

func distribute(world [][]byte, p Params, c distributorChannels, t int, heightDiff float32) [][]byte {

	// initialise slice of channels to maintain order
	// when sending tasks to worker threads
	channels := make([]chan [][]byte, p.Threads)
	for i := range channels {
		channels[i] = make(chan [][]byte)
	}

	// sets up workers for all slices
	for i := 0; i < p.Threads; i++ {
		go worker(world, p, c, t, int(float32(i)*heightDiff), int(float32(i+1)*heightDiff), channels[i])
	}


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

func saveBoard(world [][]byte, turn int, p Params, c distributorChannels) {

	filename := fmt.Sprintf("%dx%dx%d", p.ImageWidth, p.ImageHeight, turn)

	// get writePgmImage ready to recieve our world
	c.ioCommand <- ioOutput
	c.ioFilename <- filename

	// pipe the world byte by byte into ioOuput channel, for use in writePgmImage
	for i := range world {
		for j := range world[i] {
			c.ioOutput <- world[i][j]
		}
	}

	c.events <- ImageOutputComplete{CompletedTurns: turn, Filename: filename}
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

  // sends all currently alive cells to sdl
  for _, cell := range getAliveCells(world) {
		c.events <- CellFlipped{CompletedTurns: turn, Cell: cell}
  }

	// to track whether alivecells should be counted
	count := make(chan bool)
	go aliveTicker(count)

  // the height of the slices worked on by worker threads
	var heightDiff float32 = float32(p.ImageHeight) / float32(p.Threads)

	// distributes tasks for each turn depending on number of threads
	for turn = 0; turn < p.Turns; turn++ {
		world = distribute(world, p, c, turn, heightDiff)

		c.events <- TurnComplete{CompletedTurns: turn}

		// selects appropriate action based on keyboard presses/ ticker
		select {
		case key := <-c.keyPress:
			switch key {
			case 's':
				saveBoard(world, turn, p, c)
			case 'q':
				p.Turns = turn
			case 'p':
				c.events <- StateChange{turn, Paused}
				fmt.Printf("Paused on turn %d. Press p to continue... ", turn)
				for {
					if <-c.keyPress == 'p' {
						break
					}
				}
				fmt.Printf("Continuing!\n", turn)
				c.events <- StateChange{turn, Executing}
			}
		case <-count: //ticker call
			c.events <- AliveCellsCount{turn + 1, len(getAliveCells(world))}
		default:
		}

	}

	// Report the final turn being complete
	c.events <- FinalTurnComplete{turn, getAliveCells(world)}

	// save the world as a pgm file
	saveBoard(world, turn, p, c)

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{turn, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}
