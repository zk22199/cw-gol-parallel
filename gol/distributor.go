package gol

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- uint8
	ioInput    <-chan uint8
}

func executeTurn(world [][]byte) {

}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {

	// TODO: Create a 2D slice to store the world.
	newworld := make([][]byte, p.ImageWidth)
	for i := range newworld {
		newworld[i] = make([]byte, p.ImageHeight)
	}

	turn := 0

	// TODO: Execute all turns of the Game of Life.

	// initial turn is executed, store result in world

	// execute the remainder of the turns
	for i := 1; i < p.Turns; i++ {

	}

	// TODO: Report the final state using FinalTurnCompleteEvent.

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{turn, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}
