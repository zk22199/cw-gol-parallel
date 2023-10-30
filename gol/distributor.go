package gol

import (
  "uk.ac.bris.cs/gameoflife/util"
  "fmt"
)

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- uint8
	ioInput    <-chan uint8
}

func distribute(world [][]byte, p Params) [][]byte {
  channels := make([]chan [][]byte, p.Threads)
  for i := range channels {
    channels[i] = make(chan [][]byte)
  }
  heightDiff := p.ImageHeight / p.Threads  
  // sets up workers for all except last slice
  for i := 0; i < p.Threads - 1; i++ {
    go worker(world, p, i * heightDiff, (i + 1) * heightDiff, channels[i])
  }
  // sets up worker for last slice, necessary to correct
  // for inconsistencies with rounding
  go worker(world, p, (p.Threads - 1) * heightDiff, p.ImageHeight, channels[p.Threads - 1])

  var newWorld [][]byte
  for i := 0; i < p.Threads; i++ {
    thisSlice := <- channels[i]
    newWorld = append(newWorld, thisSlice...)
    //cells := getAliveCells(newWorld, p)
    //if len(cells) != 0 {
      //fmt.Println(cells)
    //}
  }
  return newWorld 
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
	world := make([][]byte, p.ImageHeight)
	for i := range world {
		world[i] = make([]byte, p.ImageWidth)
		for j := range world[i] {
			world[i][j] = <-c.ioInput
		}
	}

	turn := 0

	// TODO: Execute all turns of the Game of Life.
	for turn = 0; turn < p.Turns; turn++ {
		world = distribute(world, p)
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
