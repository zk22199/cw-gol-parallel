package gol

import(
  . "uk.ac.bris.cs/gameoflife/util"
)

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- uint8
	ioInput    <-chan uint8
}

func live(world [][]byte, i, j int) int {
  if world[i][j] == 255 {
    return 1
  } else {
    return 0
  }
}

func getLiveNeighbours(p Params, world [][]byte, i int, j int) int {
  up := i > 0
  left := j > 0
  down := i < p.ImageHeight - 1
  right := j < p.ImageWidth - 1
  res := 0

  var upCoord, leftCoord, downCoord, rightCoord int
  if up {
    upCoord = i - 1
  } else {
    upCoord = p.ImageHeight - 1
  }

  if down {
    downCoord = i + 1
  } else {
    downCoord = 0
  }


  if down {
    downCoord = i + 1
  } else {
    downCoord = 0
  }

  if left {
    leftCoord = j - 1
  } else {
    leftCoord = p.ImageWidth - 1
  }

  if left {
    leftCoord = j - 1
  } else {
    leftCoord = p.ImageWidth - 1
  }

  if right {
    rightCoord = j + 1
  } else {
    rightCoord = 0
  }

  res += live(world, upCoord, leftCoord)
  res += live(world, upCoord, j)
  res += live(world, upCoord, rightCoord)
  res += live(world, i, leftCoord)
  res += live(world, i, rightCoord)
  res += live(world, downCoord, leftCoord)
  res += live(world, downCoord, j)
  res += live(world, downCoord, rightCoord)

  return res
}

func calculateNextState(p Params, world [][]byte) [][]byte {
  var liveByte byte = 255
  var deadByte byte = 0
  res := make([][]byte , p.ImageHeight)
  for i := range res {
    res[i] = make([]byte, p.ImageWidth)
  }
  for i := 0; i < p.ImageHeight; i++ {
    for j := 0; j < p.ImageWidth; j++ {
      count := getLiveNeighbours(p, world, i, j)
      if world[i][j] == 0 {
        if count == 3 {
          res[i][j] = liveByte
        } else {
          res[i][j] = deadByte
        }
      } else {
        if count < 2 || count > 3 {
          res[i][j] = deadByte
        } else {
          res[i][j] = liveByte
        }
      }
    }
  }
  return res
}

func calculateAliveCells(p Params, world [][]byte) []Cell {
  res := []Cell{}
  for i := 0; i < p.ImageHeight; i++ {
    for j := 0; j < p.ImageWidth; j++ {
      if world[i][j] == 255 {
        c := Cell{
          X: j,
          Y: i,
        }
        res = append(res, c)
      }
    }
  }
  return res
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {


  c.ioCommand <- ioInput
  
	// TODO: Create a 2D slice to store the world.
  world := make([][]byte, p.ImageHeight)
  for row := range world {
    world[row] = make([]byte, p.ImageWidth)
  }

	// TODO: Execute all turns of the Game of Life.
  for turn:=0; turn < p.Turns; turn++ {


  
  }

  c.ioCommand <- ioInput
	// TODO: Report the final state using FinalTurnCompleteEvent.


	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{turn, Quitting}
	
	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}
