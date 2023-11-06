package gol

import (
	"uk.ac.bris.cs/gameoflife/util"
)

func worker(world [][]byte, p Params, c distributorChannels, turn int, startY, endY int, out chan [][]byte) {

	// make a new slice that is the size of the part of the
	// world to be computed by this worker
	newworld := make([][]byte, endY-startY)
	for i := range newworld {
		newworld[i] = make([]byte, p.ImageWidth)
	}

	// check each cell in the appropriate portion of the world
	for i := range world[startY:endY] {

		// variable necessary to differentiate between local slice
		// coordinates and global world coordinates
		worldIndex := startY + i

		for j := range world[worldIndex] {

			var sum uint8 = 0

			// counts the number of alive neighbours
			for m := -1; m <= 1; m++ {
				for n := -1; n <= 1; n++ {

					// the value of the currently observed cell
					// should not be added to the sum
					if m != 0 || n != 0 {
						dy := (worldIndex + m + p.ImageHeight) % p.ImageHeight
						dx := (j + n + p.ImageWidth) % p.ImageWidth

						// byte values of 255 to 1 for proccessing
						sum += (uint8(world[dy][dx]) / 255)
					}
				}
			}
			
      var oldstate uint8 = world[worldIndex][j] / 255

			// apply rules corresponding to the total surronding alive
			// cells in context for the state of the current cell
      // ensures flipping of cells only if a change is present
      if oldstate == 1 {
        if sum == 2 || sum == 3 {
				  newworld[i][j] = 255
        } else {
				  c.events <- CellFlipped{CompletedTurns: turn, Cell: util.Cell{X: j, Y: worldIndex}}
        }
      } else {
        if sum == 3 {
          newworld[i][j] = 255
          c.events <- CellFlipped{CompletedTurns: turn, Cell: util.Cell{X: j, Y: worldIndex}}
        }
      }
		}
	}
	// send computed data to the channel provided by the arguments
	out <- newworld
}
